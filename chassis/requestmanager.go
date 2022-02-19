package chassis

import (
	"context"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type RequestFunction func(requestContext *RequestContext) RequestResponse

type RequestManager struct {
	communication *RabbitCommunication

	registeredRequests  []RegisteredRequest
	unhandledHandler    RequestFunction
	requestPanicChannel chan *RequestContext

	options              *RequestManagerOptions
	serviceInfo          *ServiceInfo
	requestRabbitChannel *amqp.Channel
}

type RequestManagerOptions struct {
	requestPanicChannelSize int
	requestTimeOut          time.Duration
}

var DefaultRequestManagerOptions = &RequestManagerOptions{
	10, time.Minute,
}

type RegisteredRequest struct {
	path    string
	method  string
	action  RequestFunction
	matcher *regexp.Regexp
}

func NewRequestManager(com *RabbitCommunication, info *ServiceInfo, options *RequestManagerOptions) *RequestManager {
	if options == nil {
		options = DefaultRequestManagerOptions
	}
	return &RequestManager{
		com,
		[]RegisteredRequest{},
		nil,
		make(chan *RequestContext, options.requestPanicChannelSize),
		options,
		info,
		nil,
	}
}

func (rm *RequestManager) RegisterRequestHandler(path string, method string, action RequestFunction) error {
	matcher, err := regexp.Compile(PathToRegex(path))
	if err != nil {
		return errors.New("Failed to compile path: " + err.Error())
	}

	rm.registeredRequests = append(rm.registeredRequests, RegisteredRequest{
		path:    path,
		method:  method,
		action:  action,
		matcher: matcher,
	})
	return nil
}

var segmentMatcher = regexp.MustCompile(`{([^/{}]+)}`)

func PathToRegex(path string) string {
	regexPath := `^` + path + `(?:\?.*)?$`
	t := segmentMatcher.FindAllStringSubmatch(path, -1)
	for _, match := range t {
		group := match[1]
		var newSegment strings.Builder
		newSegment.WriteString(`(?P<`)
		if strings.Contains(group, `:`) {
			split := strings.Split(group, `:`)
			newSegment.WriteString(split[0])
			newSegment.WriteString(`>`)
			switch split[1] {
			case "int":
				newSegment.WriteString(`-?\d+`)
			case "string":
				newSegment.WriteString(`\w+`)
			case "float":
				newSegment.WriteString(`-?\d+\.\d+`)
			case "ip4":
				newSegment.WriteString(`(?:25[0-5]|2[0-4]\d|[01]\d{2}|\d{1,2})(?:.(?:25[0-5]|2[0-4]\d|[01]\d{2}|\d{1,2})){3}`)
			case "ip6":
				newSegment.WriteString(`(?:[A-Fa-f0-9]){0,4}(?: ?:? ?(?:[A-Fa-f0-9]){0,4}){0,7}`)
			case "uuid":
				newSegment.WriteString(`^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}$`)
			default:
				newSegment.WriteString(split[1])
			}

		} else {
			newSegment.WriteString(group)
			newSegment.WriteString(`>[^/?]+`)
		}
		newSegment.WriteString(`)`)
		regexPath = strings.Replace(regexPath, match[0], newSegment.String(), -1)
	}
	return regexPath
}

func (rm *RequestManager) RegisterUnhandledRequestHandler(action RequestFunction) {
	rm.unhandledHandler = action
}

func (rm *RequestManager) NewRequest(path string, method string, payload []byte, contentType string, correlationId string, replyAddress string) error {
	c := context.WithValue(context.Background(), "path", path)
	c = context.WithValue(c, "method", method)
	c = context.WithValue(c, "payload", payload)
	c = context.WithValue(c, "payload-type", contentType)

	var query url.Values
	if strings.Contains(path, "?") {
		parts := strings.Split(path, "?")
		var err error
		query, err = url.ParseQuery(parts[1])
		if err != nil {
			return err
		}
	} else {
		query = make(url.Values)
	}
	c = context.WithValue(c, "params", query)

	var handler RequestFunction
	for _, registered := range rm.registeredRequests {
		if registered.method == method && registered.matcher.MatchString(path) {

			handler = registered.action
			submatch := registered.matcher.FindAllStringSubmatch(path, -1)
			values := make(url.Values)
			for _, match := range submatch {
				for i, name := range registered.matcher.SubexpNames() {
					if i != 0 && name != "" {
						values.Add(name, match[i])
					}
				}
			}
			c = context.WithValue(c, "values", values)
			break
		}
	}
	if handler == nil {
		handler = rm.unhandledHandler
	}

	finalC, canFunc := context.WithTimeout(c, rm.options.requestTimeOut)

	go func() {
		rctx := NewRequestContext(finalC)

		defer func(context *RequestContext) {
			if r := recover(); r != nil {
				rm.requestPanicChannel <- context
				fmt.Println("Recovering from panic:", r)
			}
		}(rctx)
		defer canFunc()
		resp := handler(rctx)

		headers := amqp.Table{}
		headers["status-code"] = resp.StatusCode
		err := rm.requestRabbitChannel.Publish("requests.responses", replyAddress, false, false, amqp.Publishing{
			ContentType:   resp.ContentType,
			CorrelationId: correlationId,
			ReplyTo:       replyAddress,
			Body:          resp.Body,
			Headers:       headers,
		})

		if err != nil {
			fmt.Println("Failed to return message: " + err.Error())
		}
	}()
	return nil
}

func (rm *RequestManager) Serve() error {
	var err error
	rm.requestRabbitChannel, err = rm.communication.Connection.Channel()
	if err != nil {
		return err
	}

	err = rm.requestRabbitChannel.ExchangeDeclare("requests", "topic", false, false, false, false, nil)
	if err != nil {
		return err
	}

	serviceLabel := rm.serviceInfo.serviceName + ".requests"

	q, err := rm.requestRabbitChannel.QueueDeclare(serviceLabel, false, false, true, false, nil)
	if err != nil {
		return err
	}

	err = rm.requestRabbitChannel.QueueBind(q.Name, serviceLabel, "requests", false, nil)
	if err != nil {
		return err
	}

	msgs, err := rm.requestRabbitChannel.Consume(q.Name, serviceLabel, true, false, false, false, nil)
	if err != nil {
		return err
	}

	go rm.consumeRabbit(msgs)

	errChan := rm.requestRabbitChannel.NotifyClose(make(chan *amqp.Error))
	go func() {
		err, more := <-errChan
		if more {
			fmt.Println("Channel Closed due to: " + err.Reason)
			err := rm.Serve()
			if err != nil {
				fmt.Println("Channel reconnect failed: " + err.Error())
			}
		}
	}()

	return nil
}

func (rm *RequestManager) consumeRabbit(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		path, exists := msg.Headers["request-path"].(string)
		if !exists {
			fmt.Println("EventManager: consumeRabbit: Path not provided")
			continue
		}
		method, exists := msg.Headers["request-method"].(string)
		if !exists {
			fmt.Println("EventManager: consumeRabbit: Method not provided")
			continue
		}
		if err := rm.NewRequest(path, method, msg.Body, msg.ContentType, msg.CorrelationId, msg.ReplyTo); err != nil {
			fmt.Println("EventManager: consumeRabbit: " + err.Error())
		}
	}
}
