package main

import (
	"embedpi/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"embedpi/iotwifi"

	"github.com/bieber/barcode"
	"github.com/caarlos0/env"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
)

// ApiReturn structures a message for returned API calls.
type ApiReturn struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Payload interface{} `json:"payload"`
}

func barcodeScan() {
	webcam, err := gocv.VideoCaptureDevice(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()

	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("cannot read device %d\n", 0)
			return
		}
		if img.Empty() {
			continue
		}
		webcam.Read(&img)

		scanner := barcode.NewScanner().SetEnabledAll(true)

		imgObj, _ := img.ToImage()

		src := barcode.NewImage(imgObj)
		symbols, _ := scanner.ScanImage(src)

		for _, s := range symbols {
			data := s.Data
			fmt.Println(data)

		}
	}
}

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
	if deviceConfig.IsDev() {
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

	apiPayloadReturn := func(w http.ResponseWriter, message string, payload interface{}) {
		apiReturn := &ApiReturn{
			Status:  "OK",
			Message: message,
			Payload: payload,
		}
		ret, _ := json.Marshal(apiReturn)

		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}

	// marshallPost populates a struct with json in post body
	marshallPost := func(w http.ResponseWriter, r *http.Request, v interface{}) {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			zap.S().Error(err)
			return
		}

		defer r.Body.Close()

		decoder := json.NewDecoder(strings.NewReader(string(bytes)))

		err = decoder.Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			zap.S().Error(err)
			return
		}
	}

	// common error return from api
	retError := func(w http.ResponseWriter, err error) {
		apiReturn := &ApiReturn{
			Status:  "FAIL",
			Message: err.Error(),
		}
		ret, _ := json.Marshal(apiReturn)

		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}

	// handle /status POSTs json in the form of iotwifi.WpaConnect
	statusHandler := func(w http.ResponseWriter, r *http.Request) {

		status, err := wpacfg.Status()
		if err != nil {
			zap.S().Error(err.Error())
			return
		}

		apiPayloadReturn(w, "status", status)
	}

	// handle /connect POSTs json in the form of iotwifi.WpaConnect
	connectHandler := func(w http.ResponseWriter, r *http.Request) {
		var creds iotwifi.WpaCredentials
		marshallPost(w, r, &creds)

		zap.S().Infof("Connect Handler Got: ssid:|%s| psk:|%s|", creds.Ssid, creds.Psk)

		connection, err := wpacfg.ConnectNetwork(creds)
		if err != nil {
			zap.S().Error(err.Error())
			return
		}

		apiReturn := &ApiReturn{
			Status:  "OK",
			Message: "Connection",
			Payload: connection,
		}

		ret, err := json.Marshal(apiReturn)
		if err != nil {
			retError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}

	// scan for wifi networks
	scanHandler := func(w http.ResponseWriter, r *http.Request) {
		zap.S().Info("Got Scan")
		wpaNetworks, err := wpacfg.ScanNetworks()
		if err != nil {
			retError(w, err)
			return
		}

		apiReturn := &ApiReturn{
			Status:  "OK",
			Message: "Networks",
			Payload: wpaNetworks,
		}

		ret, err := json.Marshal(apiReturn)
		if err != nil {
			retError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}

	// kill the application
	killHandler := func(w http.ResponseWriter, r *http.Request) {
		messages <- iotwifi.CmdMessage{Id: "kill"}

		apiReturn := &ApiReturn{
			Status:  "OK",
			Message: "Killing service.",
		}
		ret, err := json.Marshal(apiReturn)
		if err != nil {
			retError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(ret)
	}

	// common log middleware for api
	logHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			staticFields := make(map[string]interface{})
			staticFields["remote"] = r.RemoteAddr
			staticFields["method"] = r.Method
			staticFields["url"] = r.RequestURI

			zap.S().Info(staticFields, "HTTP")
			next.ServeHTTP(w, r)
		})
	}

	// setup router and middleware
	r := mux.NewRouter()
	r.Use(logHandler)

	// set app routes
	r.HandleFunc("/status", statusHandler)
	r.HandleFunc("/connect", connectHandler).Methods("POST")
	r.HandleFunc("/scan", scanHandler)
	r.HandleFunc("/kill", killHandler)
	http.Handle("/", r)

	// CORS
	headersOk := handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Content-Length", "X-Requested-With", "Accept", "Origin"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	// serve http
	zap.S().Info("HTTP Listening on " + appConfig.Port)

	go func() {
		http.ListenAndServe(":"+appConfig.Port, handlers.CORS(originsOk, headersOk, methodsOk)(r))
	}()

	go func() {
		barcodeScan()
	}()

}
