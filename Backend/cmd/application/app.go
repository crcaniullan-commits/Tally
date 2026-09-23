package application

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/crcaniullan-commits/Tally/docs"
	"github.com/crcaniullan-commits/Tally/internal/auth"
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/middleware"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

const version = "0.0.1"

type Application struct {
	config Config
	db     *sql.DB
	logger *zap.SugaredLogger
}

type Config struct {
	Addr      string
	Db        DbConfig
	ApiURL    string
	JwtConfig JwtConfig
}

type JwtConfig struct {
	Secret string
	Exp    time.Duration
	Iss    string
}

type DbConfig struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

func NewApplication(cfg Config, db *sql.DB, logger *zap.SugaredLogger) *Application {
	return &Application{
		cfg,
		db,
		logger,
	}
}

func (app *Application) Run() error {

	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.ApiURL
	docs.SwaggerInfo.BasePath = "/v1"

	r := chi.NewRouter()

	middl := middleware.NewAuthMiddleware(
		errorhandler.NewErrorResponse(app.logger),
		auth.NewJWTAuthenticator(
			app.config.JwtConfig.Secret,
			app.config.JwtConfig.Iss,
			app.config.JwtConfig.Iss,
		),
		users.NewStorage(app.db),
	)

	r.Use(middleware.GlobalMiddlewares()...)

	r.Route("/v1", func(r chi.Router) {
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.Addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		auth.InitModule(r, app.db, app.logger,
			app.config.JwtConfig.Secret, app.config.JwtConfig.Iss, app.config.JwtConfig.Exp)
		r.Route("/app", func(r chi.Router) {
			r.Use(middl.AuthTokenMiddleware)
			users.InitModule(r, app.db, app.logger, middl)
		})
	})

	srv := &http.Server{
		Addr:         app.config.Addr,
		Handler:      r,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("Server has started at http://localhost/", "addr", app.config.Addr)

	return srv.ListenAndServe()
}
