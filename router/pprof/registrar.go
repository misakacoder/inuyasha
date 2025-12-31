package pprof

import (
	"github.com/gin-gonic/gin"
)

func Register(engine *gin.Engine) {
	engine.GET("/debug/pprof/*action", pprof)
}
