package db

import (
	"github.com/misakacoder/inuyasha/configs"
	"github.com/misakacoder/inuyasha/pkg/db/orm"
	innerLogger "github.com/misakacoder/inuyasha/pkg/logger"
	"github.com/misakacoder/logger"
	"gorm.io/driver/mysql"
)

var GORM *orm.Gorm

func init() {
	conf := configs.Config.Db
	GORM = orm.New(mysql.Open, conf)
	GORM.Logger = func(level string, caller string, format string, args ...any) {
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
}
