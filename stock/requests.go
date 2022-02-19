package main

import (
	"github.com/asdine/storm/v3"
	"github.com/teken/GoMicroService/chassis"
)

type requestHandlers struct {
	db *storm.DB
}

func (h requestHandlers) Get(ctx *chassis.RequestContext) chassis.RequestResponse {
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
