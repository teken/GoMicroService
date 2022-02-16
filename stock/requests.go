package main

import "github.com/teken/GoMicroService/chassis"

type requests struct {
}

func (h requests) Get(chassis.RequestContext)    {}
func (h requests) Create(chassis.RequestContext) {}
func (h requests) Update(chassis.RequestContext) {}
func (h requests) Delete(chassis.RequestContext) {}

func (h requests) Unhandled(chassis.RequestContext) {}
