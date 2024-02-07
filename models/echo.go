package models

import (
	"github.com/labstack/echo/v4"
)

type EchoRouter interface {
	// GET registers a new GET route for a path with matching handler in the router
	// with optional route-level middleware.
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// POST registers a new POST route for a path with matching handler in the
	// router with optional route-level middleware.
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// PUT registers a new PUT route for a path with matching handler in the
	// router with optional route-level middleware.
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// DELETE registers a new DELETE route for a path with matching handler in the router
	// with optional route-level middleware.
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// PATCH registers a new PATCH route for a path with matching handler in the
	// router with optional route-level middleware.
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// OPTIONS registers a new OPTIONS route for a path with matching handler in the
	// router with optional route-level middleware.
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger

	// HEAD registers a new HEAD route for a path with matching handler in the router
	// with optional route-level middleware.
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) Swagger
}

type EchoGroup interface {
	EchoRouter
	// Group automatically create tags for the swagger documentation.
	//
	// Group creates a new router group with prefix and optional group-level middleware.
	Group(prefix string, m ...echo.MiddlewareFunc) EchoGroup
}
