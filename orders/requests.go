package main

import "github.com/teken/GoMicroService/chassis"

type requestHandlers struct {
}

func (h requestHandlers) Get(*chassis.RequestContext) chassis.RequestResponse {
	return chassis.NotImplementedResponse
}
func (h requestHandlers) Create(*chassis.RequestContext) chassis.RequestResponse {
	return chassis.NotImplementedResponse
}
func (h requestHandlers) Update(*chassis.RequestContext) chassis.RequestResponse {
	return chassis.NotImplementedResponse
}
func (h requestHandlers) Delete(*chassis.RequestContext) chassis.RequestResponse {
	return chassis.NotImplementedResponse
}

func (h requestHandlers) Unhandled(*chassis.RequestContext) chassis.RequestResponse {
	return chassis.NotImplementedResponse
}
