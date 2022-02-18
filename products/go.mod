module github.com/teken/GoMicroService/products

go 1.18

require github.com/teken/GoMicroService/chassis v0.0.0

require (
	github.com/go-logr/logr v1.2.1 // indirect
	github.com/go-logr/stdr v1.2.0 // indirect
	github.com/google/uuid v1.0.0 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/streadway/amqp v1.0.0 // indirect
	go.opentelemetry.io/otel v1.3.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.3.0 // indirect
	go.opentelemetry.io/otel/sdk v1.3.0 // indirect
	go.opentelemetry.io/otel/trace v1.3.0 // indirect
	golang.org/x/sys v0.0.0-20210423185535-09eb48e85fd7 // indirect
)

replace github.com/teken/GoMicroService/chassis => ../chassis
