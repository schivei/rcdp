package main

import (
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
)

// navigateEndpoint ... Navigate page
// @Summary Navigate Page
// @Description Navigate Page
// @Tags Navigate
// @Success 200 {object} dataResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/navigateRequest [post] {object} navigateRequest
// @Body {object} navigateRequest
// @Param content body navigateRequest true "content data"
func navigateEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("navigating")
	if !allowed(w, req, http.MethodPost) {
		return
	}

	sess, err := getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	var nav navigateRequest
	err = getRequest(req, &nav)

	if throw(w, NewError(err)) {
		return
	}

	var title, url, document string

	actions := make([]chromedp.Action, 0)

	actions = append(actions, network.Enable())

	if nav.Headers != nil && len(nav.Headers) > 0 {
		headers := make(network.Headers)
		for k, v := range nav.Headers {
			headers[k] = v
		}
		actions = append(actions, network.SetExtraHTTPHeaders(headers))
	}

	if nav.Cookies != nil && len(nav.Cookies) > 0 {
		cookies := make([]*network.CookieParam, 0)
		for i := 0; i < len(nav.Cookies); i++ {
			c, _ := nav.Cookies[i].toCp()
			if c != nil {
				cookies = append(cookies, c)
			}
		}
		if len(cookies) > 0 {
			actions = append(actions, network.SetCookies(cookies))
		}
	}

	actions = append(actions, chromedp.Navigate(nav.Url))

	if nav.WaitFor != nil {
		actions = append(actions, chromedp.WaitVisible(*nav.WaitFor, chromedp.BySearch))
	}

	actions = append(actions, chromedp.Location(&url),
		chromedp.Title(&title),
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
