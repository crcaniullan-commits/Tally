package application

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/crcaniullan-commits/Tally/docs"
	"github.com/crcaniullan-commits/Tally/internal/auth"
	"github.com/crcaniullan-commits/Tally/internal/bdtx"
	"github.com/crcaniullan-commits/Tally/internal/categories"
	clavesacceso "github.com/crcaniullan-commits/Tally/internal/clavesAcceso"
	"github.com/crcaniullan-commits/Tally/internal/correo"
	"github.com/crcaniullan-commits/Tally/internal/debtors"
	"github.com/crcaniullan-commits/Tally/internal/empresas"
	"github.com/crcaniullan-commits/Tally/internal/entorno"
	"github.com/crcaniullan-commits/Tally/internal/errores"
	"github.com/crcaniullan-commits/Tally/internal/expenses"
	"github.com/crcaniullan-commits/Tally/internal/goals"
	"github.com/crcaniullan-commits/Tally/internal/incomes"
	"github.com/crcaniullan-commits/Tally/internal/intermediario"
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
	Addr         string
	BD           ConfigBD
	URLAPI       string
	ConfigJWT    ConfigJWT
	ConfigCorreo ConfigCorreo
}

type ConfigJWT struct {
	Secret string
	Exp    time.Duration
	Iss    string
}

type ConfigBD struct {
	Addr         string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}

type ConfigCorreo struct {
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
	docs.SwaggerInfo.Host = app.config.URLAPI
	docs.SwaggerInfo.BasePath = "/v1"

	r := chi.NewRouter()

	middl := intermediario.NuevoMiddlewareAutenticacion(
		errores.NuevaRespuestaErrores(app.logger),
		auth.NewJWTAuthenticator(
			app.config.ConfigJWT.Secret,
			app.config.ConfigJWT.Iss,
			app.config.ConfigJWT.Iss,
		),
		users.NuevoAlmacenUsuarios(app.db),
	)

	var smtp clavesacceso.SendEmail

	if apikey := entorno.ObtenerCadena("GMAIL_SMTP_PASSWORD", ""); apikey != "" {
		NewSmtp, err := correo.NuevoClienteGmail(
			app.config.ConfigCorreo.Username,
			app.config.ConfigCorreo.Password,
			app.config.ConfigCorreo.From,
		)

		if err != nil {
			app.logger.Errorf("Error configurando SMTP", "error", err)
		}

		smtp = NewSmtp
		app.logger.Infow("SMTP conectado con el correo", "email", app.config.ConfigCorreo.Username)
	} else {
		app.logger.Errorf("Clave de API no configurada - se usará el mock de envío de correos")
		smtp = NuevoRemitenteSimulado()
	}

	transactor := bdtx.NewTransactor(app.db)

	empresaRoute, finder := empresas.InitModule(app.db, app.logger, middl)
	emMiddl := intermediario.NuevoMiddlewareEmpresa(app.logger, finder)

	r.Use(intermediario.MiddlewaresGlobales()...)

	r.Route("/v1", func(r chi.Router) {
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.Addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))
		auth.InitModule(r, app.db, app.logger, transactor,
			app.config.ConfigJWT.Secret, app.config.ConfigJWT.Iss, app.config.ConfigJWT.Exp)
		r.Route("/aplicacion", func(r chi.Router) {
			r.Use(middl.MiddlewareTokenAutenticacion)
			users.InitModule(r, app.db, app.logger, middl)
			clavesacceso.InitModule(r, app.db, app.logger, middl, smtp)
			r.Route("/empresas", func(r chi.Router) {
				empresaRoute.Collection(r)

				r.Route("/{empresaID}", func(r chi.Router) {
					r.Use(emMiddl.EmpresaDelUsuario)

					empresaRoute.Item(r)
					incomes.InitModule(r, app.db, app.logger, middl)
					expenses.InitModule(r, app.db, app.logger, middl)
					categories.InitModule(r, app.db, app.logger, middl)
					goals.InitModule(r, app.db, app.logger, middl)
					debtors.InitModule(r, app.db, app.logger, middl)

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

	app.logger.Infow("El servidor se inició en http://localhost/", "addr", app.config.Addr)

	return srv.ListenAndServe()
}

type remitenteSimulado struct{}

func NuevoRemitenteSimulado() *remitenteSimulado {
	return &remitenteSimulado{}
}

func (e *remitenteSimulado) SendActivationCode(ctx context.Context, to, code string, expiresAt time.Time) error {
	log.Printf(
		"[MOCK EMAIL] Para: %s | Código: %s | Vence: %s",
		to, code, expiresAt.Format("02/01/2006 15:04"),
	)
	return nil
}
