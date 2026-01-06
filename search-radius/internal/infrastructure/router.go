package infrastructure

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"search-radius/global"
	driverHttp "search-radius/internal/adapters/driver/http"
	"search-radius/pkg/common/http/handler"
	"search-radius/pkg/common/http/middlewares"
)

// RouterGroup contains all routes
type RouterGroup struct {
	ShopHandler driverHttp.ShopHandler
}

// NewRouterGroup creates a new RouterGroup
func NewRouterGroup(
	shopHandler driverHttp.ShopHandler,
) *RouterGroup {
	return &RouterGroup{
		ShopHandler: shopHandler,
	}
}

// registerRoutes registers all routes
func (rg *RouterGroup) registerRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// Shop routes
	shops := api.Group("/shops")
	{
		shops.POST("/find", handler.Wrap(rg.ShopHandler.Find))
		shops.GET("/:id", handler.Wrap(rg.ShopHandler.Get))

		shops.POST("", handler.Wrap(rg.ShopHandler.Create))
		shops.PUT("/:id", handler.Wrap(rg.ShopHandler.Update))
		shops.DELETE("/:id", handler.Wrap(rg.ShopHandler.Delete))

		shops.POST("/search-radius", handler.Wrap(rg.ShopHandler.SearchByRadius))
		shops.POST("/search-radius-fast", handler.Wrap(rg.ShopHandler.SearchByRadiusFast))
	}
}

// Ping
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "OK",
		"message": "I'm running!",
	})
}

// NewEngine creates and configures the Gin engine
func NewEngine(routerGroup *RouterGroup) *gin.Engine {
	if global.Config.Server.Mode != "release" {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// middlewares
	r.Use(middlewares.CORSMiddleware)

	r.GET("/ping", Ping)

	// Register routes
	routerGroup.registerRoutes(r)

	return r
}
