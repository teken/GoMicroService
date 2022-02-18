module github.com/teken/GoMicroService/ingress

go 1.18

require (
	github.com/pborman/uuid v1.2.1
	github.com/streadway/amqp v1.0.0
	github.com/teken/GoMicroService/chassis v0.0.0
)

require (
	github.com/go-logr/logr v1.2.1 // indirect
	github.com/go-logr/stdr v1.2.0 // indirect
	github.com/google/uuid v1.0.0 // indirect
	go.opentelemetry.io/otel v1.3.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.3.0 // indirect
	go.opentelemetry.io/otel/sdk v1.3.0 // indirect
	go.opentelemetry.io/otel/trace v1.3.0 // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
)

replace github.com/teken/GoMicroService/chassis => ../chassis
