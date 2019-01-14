package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	hclog "github.com/hashicorp/go-hclog"
)

type paymentHandler struct {
	paymentGatewayURI string
	logger            hclog.Logger
}

type paymentRequest struct {
	FullName string `json:"name" valid:"alpha"`
	Number   string `json:"number" valid:"numeric"`
	CVC      string `json:"cvc" valid:"length(3,4)"`
	Expiry   string `json:"expiry" valid:"length(5,5)"`
	Type     string `json:"type"`
}

func (e paymentHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var pr paymentRequest
	if r.Body == nil {
		http.Error(rw, "Missing request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	json.NewDecoder(r.Body).Decode(&pr)

	/*
		valid, err := valid.ValidateStruct(&pr)
		if valid == false || err != nil {
			e.logger.Info("Invalid JSON payload")
			http.Error(rw, "Invalid payload", http.StatusBadRequest)
			return
		}
	*/

	data, err := json.Marshal(&pr)
	if err != nil {
		e.logger.Error("Unable to encode request body", "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	e.logger.Info("Sending request to upstream", "upstream", e.paymentGatewayURI, "request", string(data))
	resp, err := http.Post(e.paymentGatewayURI, "application/json", bytes.NewReader(data))
	if err != nil {
		e.logger.Error("Unable to send request to upstream", "upstream", e.paymentGatewayURI, "error", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		e.logger.Error("Invalid response from upstream", "upstream", e.paymentGatewayURI, "code", resp.StatusCode, "response", string(body))
		http.Error(rw, "Invalid response from payment gateway", http.StatusInternalServerError)
		return
	}

	e.logger.Info("Successfull response from upstream", "upstream", e.paymentGatewayURI)
	io.Copy(rw, resp.Body)
}
