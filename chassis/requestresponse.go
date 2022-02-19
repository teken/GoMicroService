package chassis

import "encoding/json"

type RequestResponse struct {
	StatusCode  int
	Body        []byte
	ContentType string
}

var NotImplementedResponse = StatusCodeResponse(501)
var NotFoundResponse = StatusCodeResponse(404)
var OkResponse = StatusCodeResponse(200)

func JsonResponse[T any](model T, code int) (RequestResponse, error) {
	body, err := json.Marshal(model)

	if err != nil {
		return RequestResponse{}, err
	}

	return RequestResponse{
		StatusCode:  code,
		ContentType: "application/json",
		Body:        body,
	}, nil
}

func StatusCodeResponse(code int) RequestResponse {
	return RequestResponse{
		StatusCode: code,
	}
}

func ErrorResponse(code int, message string) RequestResponse {
	err := struct {
		Message string `json:"message"`
	}{message}
	resp, _ := JsonResponse(err, 500)
	return resp
}
