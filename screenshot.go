package main

import (
	"log"
	"net/http"
	"strings"
)

// screenshotEndpoint ... Screenshot
// @Summary Screenshot
// @Description Screenshot
// @Tags Screenshot
// @Success 200 {object} dataResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/screenshot [get]
func screenshotEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("requesting screenshot")
	if !allowed(w, req, http.MethodGet) {
		return
	}

	sess, err := getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	var selector *string = nil
	if req.URL.Query().Has("selector") {
		sel := req.URL.Query().Get("selector")
		if len(strings.TrimSpace(sel)) > 0 {
			selector = &sel
		}
	}

	image, eta, e := takeScreenshot(sess, selector)

	response := NewResponse(eta, nil, nil, nil, image, e)

	if throw(w, response) {
		return
	}

	ok(w, response)
}
