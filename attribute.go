package main

import (
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"time"
)

// attributeEndpoint ... Get or Set element attribute
// @Summary Get or Set element attribute
// @Description Get or Set element attribute
// @Tags Attributes
// @Success 200 {object} attrResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/attribute [post] {object} attrRequest
// @Body {object} attrRequest
// @Param content body attrRequest true "content data"
func attributeEndpoint(w http.ResponseWriter, req *http.Request) {
	if !allowed(w, req, http.MethodPost) {
		return
	}

	log.Println("requesting attribute")

	sess, err := getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	var ar attrRequest
	err = getRequest(req, &ar)

	if throw(w, NewError(err)) {
		return
	}

	var title, url, document string

	actions := make([]chromedp.Action, 0)

	actions = append(actions, network.Enable())

	if ar.Set {
		if ar.Data == nil {
			*ar.Data = ""
		}
		actions = append(actions, chromedp.SetAttributeValue(ar.Selector, ar.Attribute, *ar.Data, chromedp.BySearch), chromedp.Sleep(100))
	}

	var value string
	var done bool
	actions = append(actions, chromedp.AttributeValue(ar.Selector, ar.Attribute, &value, &done, chromedp.BySearch))

	var sw Watcher
	var duration time.Duration

	if ar.WaitFor != nil {
		actions = append(actions, chromedp.WaitVisible(*ar.WaitFor, chromedp.BySearch))
	}

	actions = append(actions, chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.OuterHTML("html", &document))

	sw = NewStopwatch()
	err = withTimeout(ar.GetTimeout(), sess, actions...)
	duration = sw.Stop()

	response := &attrResponse{
		dataResponse: NewResponse(duration, &title, &url, &document, nil, err).(*dataResponse),
		Content:      value,
	}

	if throwCapture(w, sess, response) {
		return
	}

	ok(w, response)
}
