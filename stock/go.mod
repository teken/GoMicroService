module github.com/teken/GoMicroService/stock

go 1.18

require (
        github.com/teken/GoMicroService/chassis v0.0.0
        github.com/teken/GoMicroService/orders/events v0.0.0
        github.com/teken/GoMicroService/products/events v0.0.0
        )

replace (
        github.com/teken/GoMicroService/chassis => ../chassis
        github.com/teken/GoMicroService/orders/events => ../orders/events
        github.com/teken/GoMicroService/products/events => ../products/events
        )
