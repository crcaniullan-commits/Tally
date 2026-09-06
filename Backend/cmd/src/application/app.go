package application

import (
	"net/http"
	"time"

	"github.com/crcaniullan-commits/Tally/cmd/src/routes"
	"github.com/crcaniullan-commits/Tally/docs"
	"github.com/crcaniullan-commits/Tally/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

type Application struct {
	config Config
	store  store.Storage
	route  routes.Routes
	logger *zap.SugaredLogger
}

type Config struct {
	addr   string
	Db     DbConfig
	apiURL string
}

type DbConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func (app *Application) Run() error {

	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      app.route.Mount(),
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("Server has started", "addr", app.config.addr)

	return srv.ListenAndServe()
}

func NewApplication(cfg Config, store store.Storage, route routes.Routes, logger *zap.SugaredLogger) *Application {
	return &Application{
		cfg,
		store,
		route,
		logger,
	}
}

func NewConfig(addr string, db DbConfig, apiURL string) *Config {
	return &Config{
		addr,
		db,
		apiURL,
	}
}
