package main

import "github.com/teken/GoMicroService/chassis"

type events struct{}

func (e events) orderCreated(context chassis.EventContext) {

}

func (e events) orderCancelled(context chassis.EventContext) {

}

func (e events) orderCompleted(context chassis.EventContext) {

}

func (e events) productCreated(context chassis.EventContext) {

}

func (e events) productDeleted(context chassis.EventContext) {

}
