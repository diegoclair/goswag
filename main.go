package goswag

import (
	echoSwagger "github.com/diegoclair/goswag/internal/frameworks/echo"
	ginSwagger "github.com/diegoclair/goswag/internal/frameworks/gin"
	"github.com/diegoclair/goswag/models"
	"github.com/gin-gonic/gin"

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

type Gin interface {
	models.GinRouter
	models.GinGroup
	GenerateSwagger()
	Gin() *gin.Engine
}

// NewGin returns the interface that wraps the basic Gin methods and add the swagger methods
func NewGin(g *gin.Engine) Gin {
	return ginSwagger.NewGin(g)
}
