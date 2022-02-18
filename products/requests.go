package main

import (
	"encoding/json"
	"fmt"
	"github.com/teken/GoMicroService/chassis"
)

type requests struct {
}

type test struct {
	Test1 string `json:"test_1"`
	Test2 int    `json:"test_2"`
	Test3 test2  `json:"test_3"`
}
type test2 struct {
	Testtest1 string `json:"testtest_1"`
	Testtest2 int    `json:"testtest_2"`
}

func (h requests) GetAll(chassis.RequestContext) chassis.RequestResponse {
	t := test{
		Test1: "test1",
		Test2: 2,
		Test3: test2{Testtest1: "testtest1", Testtest2: 2},
	}
	y, _ := json.Marshal(t)
	fmt.Println(string(y[:]), t)
	resp, err := chassis.JsonResponse(t, 200)
	if err != nil {
		fmt.Println(err)
		return chassis.StatusCodeResponse(500)
	}
	return resp
}
func (h requests) Get(chassis.RequestContext) chassis.RequestResponse {
	return chassis.BlankRequestResponse
}
func (h requests) Create(chassis.RequestContext) chassis.RequestResponse {
	return chassis.BlankRequestResponse
}
func (h requests) Update(chassis.RequestContext) chassis.RequestResponse {
	return chassis.BlankRequestResponse
}
func (h requests) Delete(chassis.RequestContext) chassis.RequestResponse {
	return chassis.BlankRequestResponse
}

func (h requests) Unhandled(chassis.RequestContext) chassis.RequestResponse {
	return chassis.BlankRequestResponse
}
