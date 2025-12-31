package function

import (
	"github.com/misakacoder/kagome/errs"
	"github.com/misakacoder/logger"
)

func Sync(fn func()) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("%v", errs.GetStackTrace(err))
		}
	}()
	fn()
}

func Async(fn func()) {
	go Sync(fn)
}
