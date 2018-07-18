package main

import "net/http"

type healthHandler struct{}

func (h *healthHandler) Handle(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("OK"))
}
