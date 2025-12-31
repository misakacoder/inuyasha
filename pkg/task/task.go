package task

import (
	"github.com/misakacoder/inuyasha/pkg/function"
	"time"
)

func Register(fn func(), duration time.Duration) {
	go func() {
		ticker := time.NewTicker(duration)
		for range ticker.C {
			function.Sync(fn)
		}
	}()
}
