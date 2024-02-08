package goswag

import (
	ginWrapper "github.com/diegoclair/goswag/internal/frameworks/gin"
	"github.com/diegoclair/goswag/models"
	"github.com/gin-gonic/gin"
)

type Gin interface {
	models.GinRouter
	models.GinGroup
	GenerateSwagger()
	Gin() *gin.Engine
}

// NewGin returns the interface that wraps the basic Gin methods and add the swagger methods
func NewGin(g *gin.Engine) Gin {
	return ginWrapper.NewGin(g)
}
