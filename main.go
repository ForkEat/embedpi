package main

import (
	"embedpi/barcode"
	"embedpi/config"
	"net/http"

	"embedpi/iotwifi"

	"embedpi/api"

	"github.com/caarlos0/env"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// TODO :
// - display on eink barcode
// - add in API a way to log in into forkeat-server
// - get product with barcode
// - add menu to eink (add/remove product)

func main() {

	// Load configuration
	appConfig := config.AppConfig{}
	env.Parse(&appConfig)

	var logger *zap.Logger
	var err error

	deviceConfig, err := config.LoadCfg(appConfig.CfgUrl)
	if err != nil {
		zap.S().Error("Error to read file")
	}

	// Set log level
	if appConfig.IsDev() {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		zap.S().Error("Error to initialize logger")
	}

	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	messages := make(chan iotwifi.CmdMessage, 1)

	go iotwifi.RunWifi(messages, deviceConfig)
	wpacfg := iotwifi.NewWpaCfg(deviceConfig)

	// setup handler
	apiHandler := api.ApiHandler{WpaCfg: wpacfg, Message: messages}

	// setup router and middleware
	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			staticFields := make(map[string]interface{})
			staticFields["remote"] = r.RemoteAddr
			staticFields["method"] = r.Method
			staticFields["url"] = r.RequestURI

			zap.S().Info(staticFields, "HTTP")
			next.ServeHTTP(w, r)
		})
	})

	// set app routes
	r.HandleFunc("/status", apiHandler.StatusHandler)
	r.HandleFunc("/connect", apiHandler.ConnectHandler).Methods("POST")
	r.HandleFunc("/scan", apiHandler.ScanHandler)
	r.HandleFunc("/kill", apiHandler.KillHandler)
	http.Handle("/", r)

	// CORS
	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Content-Length", "X-Requested-With", "Accept", "Origin"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	// serve http
	zap.S().Info("HTTP Listening on " + appConfig.Port)

	go func() {
		barcode.RunScan()
	}()

	http.ListenAndServe(":"+appConfig.Port, handlers.CORS(originsOk, headersOk, methodsOk)(r))
}
