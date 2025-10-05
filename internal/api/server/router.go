package server

import (
	"github.com/wb-go/wbf/ginext"

	"github.com/avraam311/delayed-notifier/internal/api/handlers"
	"github.com/avraam311/delayed-notifier/internal/middlewares"
)

func NewRouter(handler *handlers.HandlerNotifications) *ginext.Engine {
	e := ginext.New()

	e.Use(middlewares.CORSMiddleware())
	e.Use(ginext.Logger())
	e.Use(ginext.Recovery())

	api := e.Group("/api/notify")
	{
		api.POST("/", handler.CreateNotification)
		api.GET("/:id", handler.GetNotificationStatus)
		api.DELETE("/:id", handler.DeleteNotification)
	}

	return e
}
