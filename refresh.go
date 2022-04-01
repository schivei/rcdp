package main

import (
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
)

// refreshEndpoint ... Refresh page
// @Summary Refresh Page
// @Description Refresh Page
// @Tags Refresh
// @Success 200 {object} dataResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/refreshRequest [post] {object} refreshRequest
// @Body {object} refreshRequest
// @Param content body refreshRequest true "content data"
func refreshEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("refreshing page")
	if !allowed(w, req, http.MethodPost) {
		return
	}

	sess, err := getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	var nav refreshRequest
	err = getRequest(req, &nav)

	if throw(w, NewError(err)) {
		return
	}

	var title, url, document string

	actions := make([]chromedp.Action, 0)

	actions = append(actions, network.Enable(), chromedp.Reload())

	if nav.WaitFor != nil {
		actions = append(actions, chromedp.WaitVisible(*nav.WaitFor, chromedp.BySearch))
	}

	actions = append(actions, chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.OuterHTML("html", &document))

	sw := NewStopwatch()
	err = withTimeout(nav.GetTimeout(), sess, actions...)
	value := sw.Stop()

	response := NewResponse(value, &title, &url, &document, nil, err)

	if throwCapture(w, sess, response) {
		return
	}

	ok(w, response)
}
