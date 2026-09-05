package application

import (
	"net/http"
	"time"

	"github.com/crcaniullan-commits/Tally/cmd/src/routes"
	"github.com/crcaniullan-commits/Tally/docs"
	"go.uber.org/zap"
)

const version = "0.0.1"

type Application struct {
	config Config
	route  routes.Routes
	logger *zap.SugaredLogger
}

type Config struct {
	addr   string
	apiURL string
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

func NewApplication(cfg Config, route routes.Routes, logger *zap.SugaredLogger) *Application {
	return &Application{
		cfg,
		route,
		logger,
	}
}

func NewConfig(addr string, apiURL string) *Config {
	return &Config{
		addr,
		apiURL,
	}
}
