package main

import (
	"net/http"
	"time"
)

type requestHandler struct {
	routeManager *RouteManager
}

func newRequestHandler(routeManager *RouteManager) *requestHandler {
	return &requestHandler{routeManager: routeManager}
}

func (rh *requestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	route := rh.routeManager.GetRoute(method, path)
	if route == nil {
		w.WriteHeader(404) //todo: fill out
		return
	}

	awaitResponse, err := route.SendRequest(r)
	if err != nil {
		w.WriteHeader(500) //todo: fill out
		return
	}

	select {
	case resp := <-awaitResponse:
		w.Header().Add("Content-Type", resp.ContentType)
		w.WriteHeader(resp.StatusCode)
		w.Write(resp.Body)
	case <-time.After(5 * time.Second):
		w.WriteHeader(408)
	}

}
