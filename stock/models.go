package main

type Stock struct {
	ReferenceId string `json:"reference_id,omitempty" storm:"id,index,unique"`

	ProductId        string `json:"product_id" storm:"index"`
	TotalQuantity    int    `json:"total_quantity"`
	ReservedQuantity int    `json:"reserved_quantity"`
}
