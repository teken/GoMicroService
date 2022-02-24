package main

import (
	"fmt"
	"github.com/asdine/storm/v3"
	"github.com/pborman/uuid"
	"github.com/teken/GoMicroService/chassis"
	"github.com/teken/GoMicroService/products/events"
	"strconv"
)

type requestHandlers struct {
	db *storm.DB
}

func (h requestHandlers) GetAll(ctx *chassis.RequestContext) chassis.RequestResponse {
	products := make([]Product, 1)

	pageNumber, err := strconv.Atoi(ctx.QueryParamWithDefault("page", "1"))
	if err != nil {
		fmt.Println("Failed to page number" + err.Error())
		pageNumber = 1
	}

	err = h.db.All(products, storm.Limit(100), storm.Skip((pageNumber-1)*100))
	if err != nil {
		fmt.Println("Failed to GetAll Products" + err.Error())
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	resp, err := chassis.JsonResponse(products, 200)
	if err != nil {
		fmt.Println(err)
		return chassis.StatusCodeResponse(500)
	}
	return resp
}
func (h requestHandlers) Get(ctx *chassis.RequestContext) chassis.RequestResponse {
	product := new(Product)

	referenceId := ctx.UrlParam("id")

	err := h.db.One("reference_id", referenceId, product)
	if err != nil {
		if err == storm.ErrNotFound {
			return chassis.NotFoundResponse
		}
		fmt.Println("Failed to Get Product" + err.Error())
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	resp, err := chassis.JsonResponse(product, 200)
	if err != nil {
		fmt.Println(err)
		return chassis.StatusCodeResponse(500)
	}
	return resp
}
func (h requestHandlers) Create(ctx *chassis.RequestContext) chassis.RequestResponse {
	product := new(Product)
	err := ctx.FromJson(product)
	if err != nil {
		fmt.Println("Failed to parse Product from body:" + err.Error())
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	product.ReferenceId = uuid.NewRandom().String()
	err = h.db.Save(&product)
	if err != nil {
		fmt.Println("Unable to save product")
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	err = chassis.SendJsonEvent(ctx.UserId(), events.ProductCreated, product)
	if err != nil {
		fmt.Println("Unable to send create product event")
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	resp, err := chassis.JsonResponse(product, 200)
	if err != nil {
		fmt.Println(err)
		return chassis.StatusCodeResponse(500)
	}
	return resp
}
func (h requestHandlers) BulkCreate(ctx *chassis.RequestContext) chassis.RequestResponse {
	products := make([]Product, 1)
	err := ctx.FromJson(products)
	if err != nil {
		fmt.Println("Failed to parse []Product from body:" + err.Error())
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	newIds := make([]string, len(products))
	for i, product := range products {
		newId := uuid.NewRandom().String()
		newIds[i] = newId
		product.ReferenceId = newId
		err = h.db.Save(&product)
		if err != nil {
			newIds[i] = ""
			fmt.Println("Unable to save product")
		}

		err = chassis.SendJsonEvent(ctx.UserId(), events.ProductCreated, product)
		if err != nil {
			fmt.Println("Unable to send create product event")
		}
	}
	if err != nil {
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	resp, err := chassis.JsonResponse(newIds, 200)
	if err != nil {
		fmt.Println(err)
		return chassis.StatusCodeResponse(500)
	}
	return resp
}
func (h requestHandlers) Update(ctx *chassis.RequestContext) chassis.RequestResponse {
	product := new(Product)

	referenceId := ctx.UrlParam("id")

	err := h.db.One("reference_id", referenceId, product)
	if err != nil {
		if err == storm.ErrNotFound {
			return chassis.NotFoundResponse
		}
		fmt.Println("Failed to Get Product for update" + err.Error())
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	product.ReferenceId = referenceId

	err = h.db.Update(product)
	if err != nil {
		fmt.Println("Failed to Update Product:" + err.Error())
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	err = chassis.SendJsonEvent(ctx.UserId(), events.ProductUpdated, product)
	if err != nil {
		fmt.Println("Unable to send delete product event")
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	return chassis.OkResponse
}
func (h requestHandlers) Delete(ctx *chassis.RequestContext) chassis.RequestResponse {
	product := new(Product)

	referenceId := ctx.UrlParam("id")

	err := h.db.One("reference_id", referenceId, product)
	if err != nil {
		if err == storm.ErrNotFound {
			return chassis.NotFoundResponse
		}
		fmt.Println("Failed to Get Product for deletion:" + err.Error())
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	err = h.db.DeleteStruct(product)
	if err != nil {
		fmt.Println("Failed to Delete Product:" + err.Error())
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	err = chassis.SendJsonEvent(ctx.UserId(), events.ProductDeleted, product)
	if err != nil {
		fmt.Println("Unable to send delete product event")
		return chassis.ErrorResponse(500, "Internal Server Error")
	}

	return chassis.OkResponse
}

func (h requestHandlers) Unhandled(*chassis.RequestContext) chassis.RequestResponse {
	return chassis.NotFoundResponse
}
