package main

import (
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"strings"
)

// innerHtmlEndpoint ... Get inner HTML from element
// @Summary Get inner HTML from element
// @Description Get inner HTML from element
// @Tags Inner HTML
// @Success 200 {object} innerResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/inner_html [post] {object} innerRequest
// @Body {object} innerRequest
// @Param content body innerRequest true "content data"
func innerHtmlEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("requesting html")
	var sess *session
	var request *innerRequest
	var err error
	var title, document, url, innerHtml string
	var sw Watcher

	if !allowed(w, req, http.MethodPost) {
		return
	}

	sess, err = getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	err = getRequest(req, request)
	if throw(w, NewError(err)) {
		return
	}

	actions := make([]chromedp.Action, 0)
	actions = append(actions, chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.OuterHTML("/html", &document, chromedp.BySearch),
		chromedp.InnerHTML(request.Selector, &innerHtml, chromedp.BySearch))

	if request.GetTimeout() > 0 {
		sw = NewStopwatch()
		err = withTimeout(request.GetTimeout(), sess, actions...)
	} else {
		sw = NewStopwatch()
		err = chromedp.Run(sess.cctx, actions...)
	}
	value := sw.Stop()

	var it *string = nil
	if len(strings.TrimSpace(innerHtml)) != 0 {
		it = &innerHtml
	}

	response := &innerResponse{
		Content:      it,
		dataResponse: NewResponse(value, &title, &url, &document, nil, err).(*dataResponse),
	}

	if throwCapture(w, sess, response) {
		return
	}

	ok(w, response)
}
