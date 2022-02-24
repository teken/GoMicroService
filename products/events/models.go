package events

type Product struct {
	ReferenceId string  `json:"reference_id,omitempty" storm:"id,index,unique"`
	Name        string  `json:"name,omitempty"`
	Category    string  `json:"category,omitempty"`
	SubCategory string  `json:"subcategory,omitempty"`
	Price       float64 `json:"price,omitempty"`
	SalesPrice  float64 `json:"sales_price,omitempty"`
}

func ProductFromProduct() {

}
