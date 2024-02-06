package goswag

import (
	echoSwagger "github.com/diegoclair/goswag/internal/frameworks/echo"
	"github.com/diegoclair/goswag/models"
	"github.com/labstack/echo/v4"
)

const (
	StringType = "string"
	IntType    = "int"
	NumberType = "number"
	BoolType   = "boolean"
)

type Echo interface {
	models.EchoGroup
	GenerateSwagger()
	Echo() *echo.Echo
}

// NewEcho returns the interface that wraps the basic Echo methods and add the swagger methods
func NewEcho() Echo {
	return echoSwagger.NewEcho()
}

// NewGin returns the interface that wraps the basic Gin methods and add the swagger methods
// func NewGin() Gin {
// 	return newGin()
// }
