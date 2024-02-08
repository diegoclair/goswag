package goswag

import (
	echoWrapper "github.com/diegoclair/goswag/internal/frameworks/echo"
	"github.com/diegoclair/goswag/models"

	"github.com/labstack/echo/v4"
)

type Echo interface {
	models.EchoGroup
	GenerateSwagger()
	Echo() *echo.Echo
}

// NewEcho returns the interface that wraps the basic Echo methods and add the swagger methods
func NewEcho() Echo {
	return echoWrapper.NewEcho()
}
