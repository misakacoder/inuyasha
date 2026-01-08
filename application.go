package inuyasha

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/misakacoder/inuyasha/configs"
	"github.com/misakacoder/inuyasha/consts"
	"github.com/misakacoder/inuyasha/http/resp"
	"github.com/misakacoder/inuyasha/middleware"
	innerLogger "github.com/misakacoder/inuyasha/pkg/logger"
	"github.com/misakacoder/kagome/cond"
	"github.com/misakacoder/kagome/net"
	"github.com/misakacoder/logger"
	"net/http"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	once sync.Once
	app  *application
)

type application struct {
	AppName        string
	Version        string
	BuildTime      string
	startTime      time.Time
	listeners      []configs.Listener
	middlewares    []gin.HandlerFunc
	staticHandlers []gin.HandlerFunc
	before         func()
	engine         func(engine *gin.Engine)
	after          func()
}

func New(appName string, version string, buildTime string) *application {
	once.Do(func() {
		app = &application{
			AppName:   appName,
			Version:   version,
			BuildTime: buildTime,
			startTime: time.Now(),
		}
	})
	return app
}

func (application *application) AddConfigListener(listeners ...configs.Listener) {
	application.listeners = append(application.listeners, listeners...)
}

func (application *application) AddMiddleware(handlers ...gin.HandlerFunc) {
	application.middlewares = append(application.middlewares, handlers...)
}

func (application *application) AddStaticHandler(handlers ...gin.HandlerFunc) {
	application.staticHandlers = append(application.staticHandlers, handlers...)
}

func (application *application) Before(fn func()) {
	application.before = fn
}

func (application *application) Engine(engine func(engine *gin.Engine)) {
	application.engine = engine
}

func (application *application) After(fn func()) {
	application.after = fn
}

func (application *application) Serve() {
	application.listenConfig()
	application.initLogger()
	appName := application.AppName
	logger.Info("The %s version is %s and the build time is %s", appName, application.Version, application.BuildTime)
	before := application.before
	if before != nil {
		before()
	}
	engine := application.newEngine()
	engineFunc := application.engine
	if engineFunc != nil {
		engineFunc(engine)
	}
	after := application.after
	if after != nil {
		after()
	}
	conf := configs.Config.Server
	bind := conf.Bind
	port := conf.Port
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", bind, port),
		Handler:           engine,
		ReadTimeout:       conf.ReadTimeout,
		ReadHeaderTimeout: conf.ReadHeaderTimeout,
		WriteTimeout:      conf.WriteTimeout,
		IdleTimeout:       conf.IdleTimeout,
		MaxHeaderBytes:    conf.MaxHeaderBytes,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Panic("Startup error: %s", err.Error())
		}
	}()
	banner := strings.Builder{}
	startupTime := time.Since(application.startTime)
	banner.WriteString(fmt.Sprintf("Started %s in %.2f seconds...", appName, startupTime.Seconds()))
	addresses := cond.Ternary(bind == consts.AnyAddress, net.GetLocalAddr(), []string{bind})
	for _, address := range addresses {
		banner.WriteString(fmt.Sprintf("\n - Listen on: http://%s:%d", address, port))
	}
	logger.Info(banner.String())
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Shutdown error: %s", err.Error())
	}
	logger.Info("Shutdown...")
}

func (application *application) initLogger() {
	conf := configs.Config.Log
	filename := filepath.Join(conf.Directory, fmt.Sprintf("%s.log", application.AppName))
	logger.SetLogger(innerLogger.NewDynamicLevelLogger(logger.NewSimpleLogger(filename)))
	level, _ := logger.Parse(conf.Level)
	logger.SetLevel(level)
}

func (application *application) listenConfig() {
	for _, listener := range application.listeners {
		configs.AddListener(listener.Config, listener.Reload)
	}
	configs.ListenConfig()
}

func (application *application) newEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	for _, handler := range application.staticHandlers {
		engine.Use(handler)
	}
	engine.Use(middleware.CSRF)
	engine.Use(middleware.Recovery)
	for _, handler := range application.middlewares {
		engine.Use(handler)
	}
	engine.NoRoute(resp.NotFound)
	return engine
}
