package db

import (
	"github.com/misakacoder/inuyasha/configs"
	"github.com/misakacoder/inuyasha/pkg/db/orm"
	innerLogger "github.com/misakacoder/inuyasha/pkg/logger"
	"github.com/misakacoder/logger"
	"gorm.io/gorm"
	"sync"
)

var (
	GORM *orm.Gorm
	once sync.Once
)

func Connect(dialector func(dsn string) gorm.Dialector) {
	once.Do(func() {
		conf := configs.Config.Db
		GORM = orm.New(dialector, conf)
		GORM.Logger = Logger
	})
}

func Logger(level string, caller string, format string, args ...any) {
	lvl, ok := logger.Parse(level)
	if !ok {
		lvl = logger.DEBUG
	}
	switch log := logger.GetLogger().(type) {
	case *innerLogger.DynamicLevelLogger:
		log.Push(lvl, caller, format, args...)
	case *logger.SimpleLogger:
		log.Push(lvl, caller, format, args...)
	default:
		logger.Warn("unknown logger: %T", log)
	}
}
