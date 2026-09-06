package main

import (
	"os"

	"github.com/crcaniullan-commits/Tally/cmd/src/application"
	"github.com/crcaniullan-commits/Tally/cmd/src/handler"
	"github.com/crcaniullan-commits/Tally/cmd/src/routes"
	"github.com/crcaniullan-commits/Tally/internal/db"
	"github.com/crcaniullan-commits/Tally/internal/env"
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/service"
	"github.com/crcaniullan-commits/Tally/internal/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const version = "0.0.1"

//	@title	Tally
//	@description	API para Tally, app financiera para emprendedores
//	@termsOfService http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath		/v1
//
// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						Autorization
// @description
func main() {
	cfg := application.NewConfig(
		env.GetString("ADDR", ":8080"),
		application.DbConfig{
			Addr:         env.GetString("DB_ADDR", "postgres://postgres:tally_password_env@localhost/tally?sslmode=disable"),
			MaxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			MaxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			MaxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env.GetString("EXTERNAL_URL", "localhost:8080"))

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

	db, err := db.New(
		cfg.Db.Addr,
		cfg.Db.MaxOpenConns,
		cfg.Db.MaxIdleConns,
		cfg.Db.MaxIdleTime,
	)

	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	e := errorhandler.NewErrorResponse(logger)

	store := store.NewStorage(db)
	services := service.NewService(store)
	handler := handler.NewHandler(services, e)
	router := routes.NewRoutes(handler)

	app := application.NewApplication(
		*cfg,
		store,
		*router,
		logger)

	logger.Fatal(app.Run())
}
