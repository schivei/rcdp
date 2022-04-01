package main

import (
	"log"
	"net/http"
)

// closeEndpoint ... Close browser
// @Summary Close browser
// @Description Close browser
// @Tags Close browser
// @Success 200 {object} dataResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/close [delete]
func closeEndpoint(w http.ResponseWriter, req *http.Request) {
	if !allowed(w, req, http.MethodDelete) {
		return
	}
	log.Println("closing session")
	defer deleteSession(req)
	w.WriteHeader(204)
}
