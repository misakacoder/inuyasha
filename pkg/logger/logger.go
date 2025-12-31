package logger

import (
	"fmt"
	"github.com/misakacoder/kagome/maps"
	"github.com/misakacoder/logger"
	"os"
	"runtime"
	"strings"
	"time"
)

type DynamicLevelLogger struct {
	*logger.SimpleLogger
	callerMap maps.ExpiredMap[string, logger.Level]
	tableMap  maps.ExpiredMap[string, logger.Level]
}

func NewDynamicLevelLogger(simpleLogger *logger.SimpleLogger) *DynamicLevelLogger {
	return &DynamicLevelLogger{
		SimpleLogger: simpleLogger,
		callerMap:    maps.NewExpiredMap[string, logger.Level](),
		tableMap:     maps.NewExpiredMap[string, logger.Level](),
	}
}

func (receiver *DynamicLevelLogger) Debug(message string, args ...any) {
	receiver.Push(logger.DEBUG, "", message, args...)
}

func (receiver *DynamicLevelLogger) Info(message string, args ...any) {
	receiver.Push(logger.INFO, "", message, args...)
}

func (receiver *DynamicLevelLogger) Warn(message string, args ...any) {
	receiver.Push(logger.WARN, "", message, args...)
}

func (receiver *DynamicLevelLogger) Error(message string, args ...any) {
	receiver.Push(logger.ERROR, "", message, args...)
}

func (receiver *DynamicLevelLogger) Panic(message string, args ...any) {
	receiver.Push(logger.PANIC, "", message, args...)
	os.Exit(1)
}

func (receiver *DynamicLevelLogger) Push(level logger.Level, caller string, message string, args ...any) {
	if caller == "" {
		_, file, line, _ := runtime.Caller(3)
		caller = fmt.Sprintf("%s:%d", file, line)
	}
	i := strings.LastIndex(caller, ":")
	foundLevel := false
	if value, ok := receiver.callerMap.Get(caller[0:i]); ok {
		level = value
		foundLevel = true
	}
	if !foundLevel {
		receiver.tableMap.Range(func(k string, v logger.Level) {
			for _, arg := range args {
				if foundLevel {
					return
				}
				sql, ok := arg.(string)
				if ok && strings.Contains(sql, fmt.Sprintf("`%s`", k)) {
					level = v
					foundLevel = true
				}
			}
		})
	}
	receiver.SimpleLogger.Push(level, caller, message, args...)
}

func (receiver *DynamicLevelLogger) SetCallerLevel(caller string, level logger.Level, expireTime time.Duration) {
	receiver.callerMap.PutTimeout(caller, level, expireTime)
}

func (receiver *DynamicLevelLogger) SetTableLevel(table string, level logger.Level, expireTime time.Duration) {
	receiver.tableMap.PutTimeout(table, level, expireTime)
}
