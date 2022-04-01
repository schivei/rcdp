package main

import (
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"strings"
)

// innerTextEndpoint ... Get inner Text from element
// @Summary Get inner Text from element
// @Description Get inner Text from element
// @Tags Inner Text
// @Success 200 {object} innerResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/inner_text [post] {object} innerRequest
// @Body {object} innerRequest
// @Param content body innerRequest true "content data"
func innerTextEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("requesting text")
	var sess *session
	var request *innerRequest
	var err error
	var title, document, url, innerText string
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
		chromedp.TextContent(request.Selector, &innerText, chromedp.BySearch))

	if request.GetTimeout() > 0 {
		sw = NewStopwatch()
		err = withTimeout(request.GetTimeout(), sess, actions...)
	} else {
		sw = NewStopwatch()
		err = chromedp.Run(sess.cctx, actions...)
	}
	value := sw.Stop()

	var it *string = nil
	if len(strings.TrimSpace(innerText)) != 0 {
		it = &innerText
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
