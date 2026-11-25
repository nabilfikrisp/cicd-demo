package httpserver

import "github.com/gin-gonic/gin"

// RouteRegistrar registers a module's HTTP routes.
type RouteRegistrar interface {
	RegisterRoutes(gin.IRoutes)
}

// NewRouter wires HTTP routes to their handlers.
func NewRouter(services RouteRegistrar, version, commit string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	services.RegisterRoutes(r)
	r.GET("/health", Health)
	r.GET("/version", Version(version, commit))
	return r
}
