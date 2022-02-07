package chassis

import (
	"net/http"
)

type Requests struct {
	requestManager *RequestManager
}

func (r Requests) Get(path string, action RequestFunction) error {
	return r.Request(path, http.MethodGet, action)
}
func (r Requests) Put(path string, action RequestFunction) error {
	return r.Request(path, http.MethodPut, action)
}
func (r Requests) Post(path string, action RequestFunction) error {
	return r.Request(path, http.MethodPost, action)
}
func (r Requests) Patch(path string, action RequestFunction) error {
	return r.Request(path, http.MethodPatch, action)
}
func (r Requests) Delete(path string, action RequestFunction) error {
	return r.Request(path, http.MethodDelete, action)
}
func (r Requests) Options(path string, action RequestFunction) error {
	return r.Request(path, http.MethodDelete, action)
}
func (r Requests) Request(path string, method string, action RequestFunction) error {
	return r.requestManager.RegisterRequestHandler(path, method, action)
}

func (r Requests) Unhandled(action RequestFunction) {
	r.requestManager.RegisterUnhandledRequestHandler(action)
}

func (r Requests) RequestPanicChannel() <-chan RequestContext {
	return r.requestManager.requestPanicChannel
}
