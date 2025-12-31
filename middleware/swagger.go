package middleware

import (
	"embed"
	"github.com/gin-gonic/gin"
	"github.com/misakacoder/inuyasha/configs"
)

func Swagger(embedFS embed.FS) gin.HandlerFunc {
	handler := AuthFS(NewEmbedFS(embedFS), swaggerAuth)
	return func(ctx *gin.Context) {
		if configs.Config.Swagger.Enabled {
			handler(ctx)
		} else {
			ctx.Next()
		}
	}
}

func swaggerAuth(ctx *gin.Context) bool {
	auth := configs.Config.Swagger.Auth
	if auth.Enabled {
		gin.BasicAuth(gin.Accounts{auth.Username: auth.Password})(ctx)
		return !ctx.IsAborted()
	}
	return true
}
