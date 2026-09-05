package main

import (
	"os"

	"github.com/crcaniullan-commits/Tally/cmd/src/application"
	"github.com/crcaniullan-commits/Tally/cmd/src/handler"
	"github.com/crcaniullan-commits/Tally/cmd/src/routes"
	"github.com/crcaniullan-commits/Tally/internal/env"
	"github.com/crcaniullan-commits/Tally/internal/service"
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

	services := service.NewService("database")
	handler := handler.NewHandler(services)
	router := routes.NewRoutes(handler)

	app := application.NewApplication(
		*cfg,
		*router,
		logger)

	logger.Fatal(app.Run())
}
