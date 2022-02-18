package chassis

import "encoding/json"

type RequestResponse struct {
	StatusCode  int
	Body        []byte
	ContentType string
}

var BlankRequestResponse = RequestResponse{}

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
