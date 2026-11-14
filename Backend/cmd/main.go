package main

import (
	"os"
	"time"

	"github.com/crcaniullan-commits/Tally/cmd/application"
	"github.com/crcaniullan-commits/Tally/internal/bd"
	"github.com/crcaniullan-commits/Tally/internal/entorno"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//	@title			Tally
//	@description	API para Tally, app financiera para emprendedores
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
//
// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	cfg := application.Config{
		Addr: entorno.ObtenerCadena("ADDR", ":8080"),
		BD: application.ConfigBD{
			Addr:         entorno.ObtenerCadena("DB_ADDR", "postgres://postgres:tally_password_env@localhost/tally?sslmode=disable"),
			MaxOpenConns: entorno.ObtenerEntero("DB_MAX_OPEN_CONNS", 30),
			MaxIdleConns: entorno.ObtenerEntero("DB_MAX_IDLE_CONNS", 30),
			MaxIdleTime:  entorno.ObtenerCadena("DB_MAX_IDLE_TIME", "15m"),
		},
		URLAPI: entorno.ObtenerCadena("EXTERNAL_URL", "localhost:8080"),
		ConfigJWT: application.ConfigJWT{
			Secret: entorno.ObtenerCadena("JWT_SECRET", "llaveSecreta"),
			Exp:    time.Hour * 24 * 3,
			Iss:    "Tally",
		},
		ConfigCorreo: application.ConfigCorreo{
			Username: entorno.ObtenerCadena("GMAIL_SMTP_USERNAME", ""),
			Password: entorno.ObtenerCadena("GMAIL_SMTP_PASSWORD", ""),
			From:     entorno.ObtenerCadena("GMAIL_SMTP_FROM", ""),
		},
	}

	//Logger
	loggerConfig := zap.NewProductionEncoderConfig()

	loggerConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loggerConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(loggerConfig),
		zapcore.AddSync(os.Stdout),
		zap.DebugLevel,
	)

	logger := zap.New(core).Sugar()
	defer logger.Sync()

	db, err := bd.Nuevo(
		cfg.BD.Addr,
		cfg.BD.MaxOpenConns,
		cfg.BD.MaxIdleConns,
		cfg.BD.MaxIdleTime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	app := application.NewApplication(
		cfg,
		db,
		logger,
	)

	logger.Fatal(app.Run())
}
