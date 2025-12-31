package logger

import (
	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine, middleware ...gin.HandlerFunc) {
	logger := engine.Group("/api/logger", middleware...)
	{
		logger.PUT("/level", setLevel)
	}
}
