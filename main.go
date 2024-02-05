package goswag

const (
	StringType = "string"
	IntType    = "int"
	NumberType = "number"
	BoolType   = "boolean"
)

type Swagger interface {
	Summary(summary string) Swagger
	Description(description string) Swagger
	Tags(tags ...string) Swagger
	Accept(accept ...string) Swagger
	Produce(produce ...string) Swagger
	Read(data interface{}) Swagger
	Returns(data []ReturnType) Swagger
	QueryParam(name, description, dataType string, required bool) Swagger
	HeaderParam(name, description, dataType string, required bool) Swagger
}

// NewSwaggerEcho returns the interface that wraps the basic Echo methods and add the swagger methods
func NewSwaggerEcho() Echo {
	return newSwaggerEcho()
}

// NewSwaggerGin returns the interface that wraps the basic Gin methods and add the swagger methods
// func NewSwaggerGin() Gin {
// 	return newSwaggerGin()
// }
