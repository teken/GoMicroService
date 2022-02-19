package main

import (
	"github.com/asdine/storm/v3"
	"github.com/teken/GoMicroService/chassis"
)

type eventHandlers struct {
	db *storm.DB
}

func (e eventHandlers) orderCreated(*chassis.EventContext) {}

func (e eventHandlers) orderCancelled(*chassis.EventContext) {}

func (e eventHandlers) orderCompleted(*chassis.EventContext) {}

func (e eventHandlers) productCreated(*chassis.EventContext) {}

func (e eventHandlers) productDeleted(*chassis.EventContext) {}
