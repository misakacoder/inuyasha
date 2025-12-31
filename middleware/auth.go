package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/misakacoder/inuyasha/http/req"
	"github.com/misakacoder/inuyasha/http/resp"
	"github.com/misakacoder/inuyasha/model"
	"github.com/misakacoder/inuyasha/pkg/jwt"
)

func Jwt(manager jwt.Manager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := req.BindHeader(ctx, &model.Token{})
		if token.Token == "" {
			token = req.BindQuery(ctx, token)
		}
		if token.Token == "" {
			token = req.BindForm(ctx, token)
		}
		if token.Token == "" {
			value, _ := ctx.Cookie("token")
			token.Token = value
		}
		claims, err := manager.Parse(token.Token)
		if claims == nil || err != nil {
			resp.NotLogin.Msg("token expired!").Write(ctx)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
