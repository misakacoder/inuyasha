package resp

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

var (
	OK               = Result{Code: 200, Status: http.StatusOK, Message: "ok"}
	Error            = Result{Code: 500, Status: http.StatusInternalServerError, Message: "error"}
	ParameterMissing = Result{Code: 10000, Status: http.StatusBadRequest, Message: "parameter missing"}
	ParameterError   = Result{Code: 10001, Status: http.StatusBadRequest, Message: "parameter error"}
	NotLogin         = Result{Code: 10002, Status: http.StatusUnauthorized, Message: "not login"}
	AccessDenied     = Result{Code: 10003, Status: http.StatusForbidden, Message: "access denied"}
	ResourceNotFound = Result{Code: 10003, Status: http.StatusNotFound, Message: "resource not found"}
	ServerError      = Result{Code: 10004, Status: http.StatusInternalServerError, Message: "server error"}
)

type Result struct {
	Code    int    `json:"code"`
	Status  int    `json:"-"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (result Result) Msg(message string, args ...any) Result {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return Result{Code: result.Code, Status: result.Status, Message: message}
}

func (result Result) With(v any) Result {
	return Result{Code: result.Code, Status: result.Status, Message: result.Message, Data: v}
}

func (result Result) Write(ctx *gin.Context) {
	ctx.JSON(result.Status, result)
}

func NotFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, ResourceNotFound)
}
