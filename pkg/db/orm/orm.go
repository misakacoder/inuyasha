package orm

import (
	"errors"
	"fmt"
	"github.com/misakacoder/kagome/cond"
	"github.com/misakacoder/kagome/errs"
	"github.com/misakacoder/kagome/str"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"time"
)

const (
	defaultMaxIdleConn     = 16
	defaultMaxOpenConn     = 32
	defaultConnMaxIdleTime = 30 * time.Minute
	defaultConnMaxLifetime = time.Hour
	defaultSlowSqlTime     = 100 * time.Millisecond
	tableOptionFormat      = "engine=InnoDB default charset=utf8mb4 collate=utf8mb4_bin comment='%s'"
	tablePartitionFormat   = "partition by %s %s"
)

type Tabler interface {
	TableComment() string
}

type Config struct {
	DSN             string        `json:"dsn"`
	MaxIdleConn     int           `yaml:"maxIdleConn"`
	MaxOpenConn     int           `yaml:"maxOpenConn"`
	ConnMaxLifeTime time.Duration `yaml:"connMaxLifeTime"`
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime"`
	SlowSqlTime     time.Duration `yaml:"slowSqlTime"`
	PrintSql        bool          `yaml:"printSql"`
}

type Gorm struct {
	*gorm.DB
	Logger         func(level string, caller string, format string, args ...any)
	namingStrategy schema.NamingStrategy
}

func (orm *Gorm) Printf(format string, args ...any) {
	level := "DEBUG"
	caller := ""
	if orm.Logger != nil && len(args) >= 2 {
		caller = args[0].(string)
		args = args[1:]
		index := strings.Index(format, "%s")
		format = strings.ReplaceAll(format, "\n", " ")[index+2:]
		format = strings.TrimPrefix(format, " ")
		switch tp := args[0].(type) {
		case error:
			if errors.Is(tp, gorm.ErrRecordNotFound) {
				level = "DEBUG"
				format = format[strings.Index(format, " ")+1:]
				args = args[1:]
			} else {
				level = "ERROR"
			}
		case string:
			if strings.Contains(tp, "SLOW SQL") {
				level = "WARN"
			}
		}
	}
	orm.Logger(level, caller, format, args...)
}

func (orm *Gorm) Transaction(fn func(tx *gorm.DB) error) {
	tx := orm.Begin()
	defer func() {
		if err := recover(); err != nil {
			tx.Rollback()
			panic(err)
		}
	}()
	err := fn(tx)
	if err != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}
}

func (orm *Gorm) TableName(model any) string {
	if tabler, ok := model.(schema.Tabler); ok {
		return tabler.TableName()
	}
	tp := reflect.TypeOf(model)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	return orm.namingStrategy.TableName(tp.Name())
}

func (orm *Gorm) AutoMigrate(models []any) {
	for _, model := range models {
		if partition, ok := model.(Partition); ok {
			parts := partition.Parts()
			for _, part := range parts {
				orm.autoMigrate(part)
			}
		} else {
			orm.autoMigrate(model)
		}
	}
}

func (orm *Gorm) autoMigrate(model any) {
	name := orm.TableName(model)
	comment := name
	if tabler, ok := model.(Tabler); ok {
		comment = tabler.TableComment()
	}
	tableOptions := fmt.Sprintf(tableOptionFormat, comment)
	if partition, ok := model.(Partition); ok {
		tp := partition.Type()
		strategy := partition.Strategy()
		if str.NoneBlank(tp, strategy) {
			tableOptions = tableOptions + " " + fmt.Sprintf(tablePartitionFormat, tp, strategy)
		}
	}
	_ = orm.Table(name).Set("gorm:table_options", tableOptions).AutoMigrate(model)
}

func New(dialector func(dsn string) gorm.Dialector, config Config) *Gorm {
	if dialector == nil {
		panic("dialector is nil")
	}
	dsn := config.DSN
	if str.NonBlank(dsn) {
		orm := &Gorm{}
		namingStrategy := schema.NamingStrategy{
			SingularTable: true,
		}
		orm.namingStrategy = namingStrategy
		gormConfig := &gorm.Config{
			SkipDefaultTransaction:                   true,
			NamingStrategy:                           namingStrategy,
			PrepareStmt:                              true,
			DisableForeignKeyConstraintWhenMigrating: true,
			QueryFields:                              true,
			CreateBatchSize:                          1000,
			TranslateError:                           true,
		}
		if config.PrintSql {
			slowSqlTime := config.SlowSqlTime
			sqlLogger := logger.New(orm, logger.Config{
				SlowThreshold:             cond.Ternary(slowSqlTime <= 0, defaultSlowSqlTime, slowSqlTime),
				Colorful:                  true,
				IgnoreRecordNotFoundError: true,
				LogLevel:                  logger.Info,
			})
			gormConfig.Logger = sqlLogger
		}
		gormDB, err := gorm.Open(dialector(dsn), gormConfig)
		errs.Panic(err)
		db, err := gormDB.DB()
		errs.Panic(err)
		maxIdleConn := config.MaxIdleConn
		maxOpenConn := config.MaxOpenConn
		connMaxLifeTime := config.ConnMaxLifeTime
		connMaxIdleTime := config.ConnMaxIdleTime
		db.SetMaxIdleConns(cond.Ternary(maxIdleConn <= 0, defaultMaxIdleConn, maxIdleConn))
		db.SetMaxOpenConns(cond.Ternary(maxOpenConn <= 0, defaultMaxOpenConn, maxOpenConn))
		db.SetConnMaxLifetime(cond.Ternary(connMaxLifeTime <= 0, defaultConnMaxLifetime, connMaxLifeTime))
		db.SetConnMaxIdleTime(cond.Ternary(connMaxIdleTime <= 0, defaultConnMaxIdleTime, connMaxIdleTime))
		orm.DB = gormDB
		return orm
	}
	panic("dsn is empty")
}
