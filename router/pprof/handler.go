package pprof

import (
	"github.com/gin-gonic/gin"
	"github.com/misakacoder/inuyasha/configs"
	"github.com/misakacoder/inuyasha/http/resp"
	"net/http"
	_ "net/http/pprof"
)

func pprof(ctx *gin.Context) {
	conf := configs.Config.Pprof
	if conf.Enabled {
		gin.BasicAuth(gin.Accounts{conf.Username: conf.Password})(ctx)
		if !ctx.IsAborted() {
			http.DefaultServeMux.ServeHTTP(ctx.Writer, ctx.Request)
		}
	} else {
		resp.NotFound(ctx)
	}
}
