package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/misakacoder/inuyasha/consts"
	"github.com/misakacoder/inuyasha/http/resp"
	"github.com/misakacoder/kagome/cond"
	"github.com/misakacoder/kagome/errs"
	"github.com/misakacoder/logger"
	"net/http"
	"time"
)

func NetWork(ctx *gin.Context) {
	start := time.Now()
	ctx.Next()
	duration := time.Since(start)
	request := ctx.Request
	logger.Info("%s %s %s %d %dms", clientIP(ctx), request.Method, request.URL.Path, ctx.Writer.Status(), duration.Milliseconds())
}

func CSRF(ctx *gin.Context) {
	method := ctx.Request.Method
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD")
	ctx.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token")
	ctx.Header("Access-Control-Expose-Headers", "Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type, Content-Length")
	ctx.Header("Access-Control-Allow-Credentials", "true")
	if method == "OPTIONS" {
		ctx.AbortWithStatus(http.StatusOK)
	}
	ctx.Next()
}

func Recovery(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			serverError := resp.ServerError.With(nil)
			printError := true
			switch errType := err.(type) {
			case error:
				serverError.Data = errType.Error()
			case resp.Result:
				serverError = errType
				printError = false
			case *resp.Result:
				serverError = *errType
				printError = false
			case string:
				serverError.Data = errType
			default:
				serverError.Data = fmt.Sprintf("%v", errType)
			}
			if printError {
				logger.Error("%v", errs.GetStackTrace(err))
			}
			serverError.Write(ctx)
			ctx.Abort()
		}
	}()
	ctx.Next()
}

func Panic(err error) {
	if err != nil {
		panic(resp.Error.Msg(err.Error()))
	}
}

func clientIP(ctx *gin.Context) string {
	ip := ctx.ClientIP()
	return cond.Ternary(ip == "::1", consts.Localhost, ip)
}
