package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"time"
)

// typeEndpoint ... Type value
// @Summary Type value
// @Description Type value
// @Tags Type value
// @Success 200 {object} dataResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/type [post] {object} typeRequest
// @Body {object} typeRequest
// @Param content body typeRequest true "content data"
func typeEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("triggering keyboard event")
	if !allowed(w, req, http.MethodPost) {
		return
	}

	sess, err := getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	var tr typeRequest
	err = getRequest(req, &tr)

	if throw(w, NewError(err)) {
		return
	}

	var title, url, document string

	actions := make([]chromedp.Action, 0)

	actions = append(actions, network.Enable())
	var sw Watcher
	var duration time.Duration

	if tr.Force {
		sw = NewStopwatch()
		err = withTimeout(tr.GetTimeout(), sess, network.Enable(),
			chromedp.Clear(tr.Selector, chromedp.BySearch),
			chromedp.SendKeys(tr.Selector, tr.Content, chromedp.BySearch),
			chromedp.Sleep(100))
		duration = sw.Stop()

		if throwCapture(w, sess, NewErrorTime(err, duration)) {
			return
		}

		var typedValue string
		sw.Continue()
		err = withTimeout(tr.GetTimeout(), sess,
			chromedp.Value(tr.Selector, &typedValue, chromedp.BySearch))
		duration = sw.Stop()

		if throwCapture(w, sess, NewErrorTime(err, duration)) {
			return
		}

		if typedValue != tr.NormalizedContent {
			sw.Continue()
			err = withTimeout(tr.GetTimeout(), sess,
				chromedp.Clear(tr.Selector, chromedp.BySearch),
				chromedp.SendKeys(tr.Selector, kb.ArrowDown, chromedp.BySearch),
				chromedp.QueryAfter(tr.Selector, func(ctx context.Context, execCtx runtime.ExecutionContextID, nodes ...*cdp.Node) error {
					if len(nodes) < 1 {
						return fmt.Errorf("selector %q did not return any nodes", tr.Selector)
					}

					return chromedp.Run(ctx, chromedp.MouseClickNode(nodes[0], chromedp.ClickCount(3)))
				}, chromedp.BySearch, chromedp.NodeVisible),
				chromedp.Sleep(100),
				chromedp.SetValue(tr.Selector, tr.Content, chromedp.BySearch),
				chromedp.SendKeys(tr.Selector, kb.ArrowDown, chromedp.BySearch),
				chromedp.Sleep(100),
				chromedp.Value(tr.Selector, &typedValue, chromedp.BySearch),
			)
			duration = sw.Stop()

			if throwCapture(w, sess, NewErrorTime(err, duration)) {
				return
			}

			if typedValue != tr.NormalizedContent && throwCapture(w, sess, NewErrorTime(errors.New("input element is not accepting content"), duration)) {
				return
			}
		}
	} else {
		var typedValue string
		sw = NewStopwatch()
		err = withTimeout(tr.GetTimeout(), sess, chromedp.SetValue(tr.Selector, tr.Content, chromedp.BySearch),
			chromedp.Sleep(100),
			chromedp.Value(tr.Selector, &typedValue, chromedp.BySearch))
		duration = sw.Stop()

		if throwCapture(w, sess, NewErrorTime(err, duration)) {
			return
		}

		if typedValue != tr.NormalizedContent && throwCapture(w, sess, NewErrorTime(errors.New("input element is not accepting content"), duration)) {
			return
		}
	}

	if tr.WaitFor != nil {
		actions = append(actions, chromedp.WaitVisible(*tr.WaitFor, chromedp.BySearch))
	}

	actions = append(actions, chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.OuterHTML("html", &document))

	sw.Continue()
	err = withTimeout(tr.GetTimeout(), sess, actions...)
	duration = sw.Stop()

	response := NewResponse(duration, &title, &url, &document, nil, err)

	if throwCapture(w, sess, response) {
		return
	}

	ok(w, response)
}
