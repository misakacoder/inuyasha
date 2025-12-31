package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/misakacoder/inuyasha/http/req"
	"github.com/misakacoder/inuyasha/http/resp"
	"github.com/misakacoder/inuyasha/model"
	innerLogger "github.com/misakacoder/inuyasha/pkg/logger"
	"github.com/misakacoder/logger"
	"time"
)

// setLevel          godoc
// @Tags             日志
// @Summary          修改
// @Router           /api/logger/level [put]
// @Param            ro body model.LogLevel false "参数"
// @Accept           json
// @Produce          json
// @Success          200 {object} resp.Result
func setLevel(ctx *gin.Context) {
	ro := req.Bind(ctx, &model.LogLevel{})
	caller := ro.Caller
	table := ro.Table
	level, ok := logger.Parse(ro.Level)
	duration, err := time.ParseDuration(ro.Time)
	if !ok {
		resp.Error.Msg("错误的日志级别").Write(ctx)
	} else if err != nil {
		resp.Error.Msg("错误的过期时间").Write(ctx)
	} else {
		dynamicLogger, isDynamicLogger := logger.GetLogger().(*innerLogger.DynamicLevelLogger)
		if isDynamicLogger {
			if caller != "" {
				dynamicLogger.SetCallerLevel(caller, level, duration)
			}
			if table != "" {
				dynamicLogger.SetTableLevel(table, level, duration)
			}
		}
		resp.OK.Write(ctx)
	}
}
