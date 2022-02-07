package chassis

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
	"time"
)

func TestPathToRegex(t *testing.T) {
	assert.Equal(t, `^/products/(?P<key>[^/?]+)(?:(?:\?)(?:[a-z0-9[]=])*)?$` ,pathToRegex("/products/{key}"))
	assert.Equal(t, `^/articles/(?P<category>[^/?]+)/(?:(?:\?)(?:[a-z0-9[]=])*)?$`, pathToRegex("/articles/{category}/"))
	assert.Equal(t, `^/articles/(?P<category>[^/?]+)/(?P<id>[0-9]+)(?:(?:\?)(?:[a-z0-9[]=])*)?$`, pathToRegex("/articles/{category}/{id:[0-9]+}"))
}

func TestNewRequest(t *testing.T) {

	manager := NewRequestManager(nil)

	go func() {
		requestContext := <-manager.requestPanicChannel
		fmt.Println("panic called")
		assert.Equal(t, "/panic", requestContext.Value("path").(string))
		assert.Equal(t, "get", requestContext.Value("method").(string))
	}()


	assert.NoError(t, manager.RegisterRequestHandler("/products/{key}", "get", func (requestContext RequestContext) {
		fmt.Println("callback 1 called")
		assert.Equal(t, "/products/908", requestContext.Value("path").(string))
		assert.Equal(t, "get", requestContext.Value("method").(string))
		urlValues := requestContext.Value("values").(url.Values)
		assert.Equal(t, "908", urlValues.Get("key"))
	}))
	assert.NoError(t, manager.RegisterRequestHandler("/type/{key}", "get", func (requestContext RequestContext) {
		fmt.Println("callback 2 called")
		assert.Equal(t, "/type/908?t=1", requestContext.Value("path").(string))
		assert.Equal(t, "get", requestContext.Value("method").(string))
		urlValues := requestContext.Value("values").(url.Values)
		assert.Equal(t, "908", urlValues.Get("key"))
		params := requestContext.Value("params").(url.Values)
		assert.Equal(t, "1", params.Get("t"))
	}))
	assert.NoError(t, manager.RegisterRequestHandler("/articles/{category}", "get", func(requestContext RequestContext) {
		fmt.Println("callback 3 called")
		assert.Equal(t, "/articles/computers", requestContext.Value("path").(string))
		assert.Equal(t, "get", requestContext.Value("method").(string))
		urlValues := requestContext.Value("values").(url.Values)
		assert.Equal(t, "computers", urlValues.Get("category"))
	}))
	assert.NoError(t, manager.RegisterRequestHandler("/articles/{category}/{id:[0-9]+}", "get", func(requestContext RequestContext) {
		fmt.Println("callback 4 called")
		assert.Equal(t, "/articles/computers/901", requestContext.Value("path").(string))
		assert.Equal(t, "get", requestContext.Value("method").(string))
		urlValues := requestContext.Value("values").(url.Values)
		assert.Equal(t, "computers", urlValues.Get("category"))
		assert.Equal(t, "901", urlValues.Get("id"))
	}))

	assert.NoError(t, manager.RegisterRequestHandler("/panic", "get", func (requestContext RequestContext) {
		panic("test")
	}))

	manager.RegisterUnhandledRequestHandler(func(requestContext RequestContext) {
		fmt.Println("callback unhandled called")
		assert.Equal(t, "/unhandled/path", requestContext.Value("path").(string))
		assert.Equal(t, "get", requestContext.Value("method").(string))
	})

	assert.NoError(t, manager.NewRequest("/products/908", "get", []byte {}))
	assert.NoError(t, manager.NewRequest("/type/908?t=1", "get", []byte {}))
	assert.NoError(t, manager.NewRequest("/articles/computers", "get", []byte {}))
	assert.NoError(t, manager.NewRequest("/articles/computers/901", "get", []byte {}))
	assert.NoError(t, manager.NewRequest("/unhandled/path", "get", []byte {}))
	assert.NoError(t, manager.NewRequest("/panic", "get", []byte {}))
	<-time.After(time.Second)
}
