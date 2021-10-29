package api

import (
	"embedpi/iotwifi"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type ApiHandler struct {
	WpaCfg  *iotwifi.WpaCfg
	Message chan iotwifi.CmdMessage
}

// ApiReturn structures a message for returned API calls.
type ApiReturn struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Payload interface{} `json:"payload"`
}

func apiPayloadReturn(w http.ResponseWriter, message string, payload interface{}) {
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
func marshallPost(w http.ResponseWriter, r *http.Request, v interface{}) {
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
func retError(w http.ResponseWriter, err error) {
	apiReturn := &ApiReturn{
		Status:  "FAIL",
		Message: err.Error(),
	}
	ret, _ := json.Marshal(apiReturn)

	w.Header().Set("Content-Type", "application/json")
	w.Write(ret)
}

// handle /status POSTs json in the form of iotwifi.WpaConnect
func (apiHandler *ApiHandler) StatusHandler(w http.ResponseWriter, r *http.Request) {

	status, err := apiHandler.WpaCfg.Status()
	if err != nil {
		zap.S().Error(err.Error())
		return
	}

	apiPayloadReturn(w, "status", status)
}

// handle /connect POSTs json in the form of iotwifi.WpaConnect
func (apiHandler *ApiHandler) ConnectHandler(w http.ResponseWriter, r *http.Request) {
	var creds iotwifi.WpaCredentials
	marshallPost(w, r, &creds)

	zap.S().Infof("Connect Handler Got: ssid:|%s| psk:|%s|", creds.Ssid, creds.Psk)

	connection, err := apiHandler.WpaCfg.ConnectNetwork(creds)
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
func (apiHandler *ApiHandler) ScanHandler(w http.ResponseWriter, r *http.Request) {
	zap.S().Info("Got Scan")
	wpaNetworks, err := apiHandler.WpaCfg.ScanNetworks()
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
func (apiHandler *ApiHandler) KillHandler(w http.ResponseWriter, r *http.Request) {
	apiHandler.Message <- iotwifi.CmdMessage{Id: "kill"}

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
