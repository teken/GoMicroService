module github.com/teken/GoMicroService/products

go 1.18

require (
        github.com/teken/GoMicroService/chassis v0.0.0
        )

replace (
        github.com/teken/GoMicroService/chassis => ../chassis
        )
