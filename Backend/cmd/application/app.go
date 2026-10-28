package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/crcaniullan-commits/Tally/docs"
	accesskeys "github.com/crcaniullan-commits/Tally/internal/accessKeys"
	"github.com/crcaniullan-commits/Tally/internal/auth"
	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/email"
	"github.com/crcaniullan-commits/Tally/internal/empresas"
	"github.com/crcaniullan-commits/Tally/internal/env"
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/incomes"
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
	Addr        string
	Db          DbConfig
	ApiURL      string
	JwtConfig   JwtConfig
	EmailConfig EmailConfig
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

type EmailConfig struct {
	Username string
	Password string
	From     string
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

	var smtp accesskeys.SendEmail

	if apikey := env.GetString("GMAIL_SMTP_PASSWORD", ""); apikey != "" {
		NewSmtp, err := email.NewGmailClient(
			app.config.EmailConfig.Username,
			app.config.EmailConfig.Password,
			app.config.EmailConfig.From,
		)

		if err != nil {
			app.logger.Errorf("Error configurando SMTP", "error", err)
		}

		smtp = NewSmtp
		app.logger.Infow("SMTP conectado con email", "email", app.config.EmailConfig.Username)
	} else {
		app.logger.Errorf("Api key no configurada - usando mock de envío de correos")
		smtp = NewMockSender()
	}

	transactor := dbtx.NewTransactor(app.db)

	empresaRoute, finder := empresas.InitModule(app.db, app.logger, middl)
	emMiddl := middleware.NewEmpresaMiddleware(app.logger, finder)

	r.Use(middleware.GlobalMiddlewares()...)

	r.Route("/v1", func(r chi.Router) {
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.Addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		auth.InitModule(r, app.db, app.logger, transactor,
			app.config.JwtConfig.Secret, app.config.JwtConfig.Iss, app.config.JwtConfig.Exp)
		r.Route("/app", func(r chi.Router) {
			r.Use(middl.AuthTokenMiddleware)
			users.InitModule(r, app.db, app.logger, middl)
			accesskeys.InitModule(r, app.db, app.logger, middl, smtp)
			r.Route("/empresas", func(r chi.Router) {
				empresaRoute.Collection(r)

				r.Route("/{empresaID}", func(r chi.Router) {
					r.Use(emMiddl.EmpresaOfUser)

					empresaRoute.Item(r)
					incomes.InitModule(r, app.db, app.logger, middl)

				})
			})
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

type emailMock struct{}

func NewMockSender() *emailMock {
	return &emailMock{}
}

func (e *emailMock) SendActivationCode(ctx context.Context, to, code string, expiresAt time.Time) error {
	log.Printf(
		"[MOCK EMAIL] Para: %s | Código: %s | Vence: %s",
		to, code, expiresAt.Format("02/01/2006 15:04"),
	)
	return nil
}
