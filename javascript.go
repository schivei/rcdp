package main

import (
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"time"
)

// javascriptEndpoint ... Execute script
// @Summary Execute script
// @Description Execute script
// @Tags Javascript
// @Success 200 {object} jsResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/javascript [post] {object} jsRequest
// @Body {object} jsRequest
// @Param content body jsRequest true "content data"
func javascriptEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("executing script")
	if !allowed(w, req, http.MethodPost) {
		return
	}

	sess, err := getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	var jr jsRequest
	err = getRequest(req, &jr)

	if throw(w, NewError(err)) {
		return
	}

	var title, url, document string

	actions := make([]chromedp.Action, 0)

	actions = append(actions, network.Enable())

	var res interface{}

	script := jr.Script

	if script == "" && throw(w, NewError(errors.New("undefined script"))) {
		return
	}

	actions = append(actions, chromedp.Evaluate(script, &res, chromedp.EvalAsValue, chromedp.EvalIgnoreExceptions))

	var sw Watcher
	var duration time.Duration

	if jr.WaitFor != nil {
		actions = append(actions, chromedp.WaitVisible(*jr.WaitFor, chromedp.BySearch))
	}

	actions = append(actions, chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.OuterHTML("html", &document))

	sw = NewStopwatch()
	err = withTimeout(jr.GetTimeout(), sess, actions...)
	duration = sw.Stop()

	response := &jsResponse{
		dataResponse: NewResponse(duration, &title, &url, &document, nil, err).(*dataResponse),
		Content:      res,
	}

	if throwCapture(w, sess, response) {
		return
	}

	ok(w, response)
}
