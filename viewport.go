package main

import (
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"time"
)

type viewportRequest struct {
	dataRequest
	Width  int64 `json:"width"`
	Height int64 `json:"height"`
}

// viewportEndpoint ... Window size
// @Summary Window size
// @Description Window size
// @Tags Window size
// @Success 200 {object} dataResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/viewport [post] {object} viewportRequest
// @Body {object} viewportRequest
// @Param content body viewportRequest true "content data"
func viewportEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("defining window size")
	if !allowed(w, req, http.MethodPost) {
		return
	}

	sess, err := getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	var r viewportRequest
	err = getRequest(req, &r)
	if throw(w, NewError(err)) {
		return
	}

	var duration time.Duration
	var title, url, document *string

	if (r.Width <= 0 || r.Height <= 0) && throw(w, NewError(errors.New("invalid viewport size"))) {
		return
	}

	if r.Timeout > 0 {
		sw := NewStopwatch()
		err = withTimeout(r.Timeout, sess, chromedp.EmulateViewport(r.Width, r.Height),
			chromedp.Title(title),
			chromedp.Location(url),
			chromedp.OuterHTML("/html", document, chromedp.BySearch))
		duration = sw.Stop()
	} else {
		sw := NewStopwatch()
		err = chromedp.Run(sess.cctx, chromedp.EmulateViewport(r.Width, r.Height),
			chromedp.Title(title),
			chromedp.Location(url),
			chromedp.OuterHTML("/html", document, chromedp.BySearch))
		duration = sw.Stop()
	}

	response := NewResponse(duration, title, url, document, nil, err)

	if throwCapture(w, sess, response) {
		return
	}

	ok(w, response)
}
