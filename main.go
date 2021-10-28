// IoT Wifi Management

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"main/config"
	"net/http"
	"os"
	"strings"

	"github.com/bhoriuchi/go-bunyan/bunyan"
	"github.com/bieber/barcode"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/txn2/txwifi/iotwifi"
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

	logConfig := bunyan.Config{
		Name:   "txwifi",
		Stream: os.Stdout,
		Level:  bunyan.LogLevelDebug,
	}

	blog, err := bunyan.CreateLogger(logConfig)
	if err != nil {
		panic(err)
	}

	blog.Info("Starting IoT Wifi...")

	messages := make(chan iotwifi.CmdMessage, 1)

	cfgUrl := config.SetEnvIfEmpty("IOTWIFI_CFG", "cfg/wificfg.json")
	port := config.SetEnvIfEmpty("IOTWIFI_PORT", "8080")

	go iotwifi.RunWifi(blog, messages, cfgUrl)
	wpacfg := iotwifi.NewWpaCfg(blog, cfgUrl)

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
			blog.Error(err)
			return
		}

		defer r.Body.Close()

		decoder := json.NewDecoder(strings.NewReader(string(bytes)))

		err = decoder.Decode(&v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			blog.Error(err)
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
			blog.Error(err.Error())
			return
		}

		apiPayloadReturn(w, "status", status)
	}

	// handle /connect POSTs json in the form of iotwifi.WpaConnect
	connectHandler := func(w http.ResponseWriter, r *http.Request) {
		var creds iotwifi.WpaCredentials
		marshallPost(w, r, &creds)

		blog.Info("Connect Handler Got: ssid:|%s| psk:|%s|", creds.Ssid, creds.Psk)

		connection, err := wpacfg.ConnectNetwork(creds)
		if err != nil {
			blog.Error(err.Error())
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
		blog.Info("Got Scan")
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

			blog.Info(staticFields, "HTTP")
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
	blog.Info("HTTP Listening on " + port)

	go func() {
		http.ListenAndServe(":"+port, handlers.CORS(originsOk, headersOk, methodsOk)(r))
	}()

	go func() {
		barcodeScan()
	}()

}
