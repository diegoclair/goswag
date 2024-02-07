package models

import "github.com/gin-gonic/gin"

type GinRouter interface {
	// Handle registers a new request handle and middleware with the given path and method.
	// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
	// See the example code in GitHub.
	//
	// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
	// functions can be used.
	//
	// This function is intended for bulk loading and to allow the usage of less
	// frequently used, non-standardized or custom methods (e.g. for internal
	// communication with a proxy).
	Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) Swagger

	// GET is a shortcut for router.Handle("GET", path, handlers).
	GET(path string, h ...gin.HandlerFunc) Swagger

	// POST is a shortcut for router.Handle("POST", path, handlers).
	POST(path string, h ...gin.HandlerFunc) Swagger

	// PUT is a shortcut for router.Handle("PUT", path, handlers).
	PUT(path string, h ...gin.HandlerFunc) Swagger

	// DELETE is a shortcut for router.Handle("DELETE", path, handlers).
	DELETE(path string, h ...gin.HandlerFunc) Swagger

	// PATCH is a shortcut for router.Handle("PATCH", path, handlers).
	PATCH(path string, h ...gin.HandlerFunc) Swagger

	// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handlers).
	OPTIONS(path string, h ...gin.HandlerFunc) Swagger

	// HEAD is a shortcut for router.Handle("HEAD", path, handlers).
	HEAD(path string, h ...gin.HandlerFunc) Swagger
}

type GinGroup interface {
	// Group automatically create tags for the swagger documentation.
	//
	// Group creates a new router group with prefix and optional group-level middleware.
	Group(prefix string, h ...gin.HandlerFunc) GinRouter
}
