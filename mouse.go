package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

// mouseEndpoint ... Mouse Event
// @Summary Mouse Event
// @Description Mouse Event
// @Tags Mouse Event
// @Success 200 {object} dataResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/mouse [post] {object} mouseRequest
// @Body {object} mouseRequest
// @Param content body mouseRequest true "content data"
func mouseEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("triggring mouse event")
	if !allowed(w, req, http.MethodPost) {
		return
	}

	sess, err := getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	var mr mouseRequest
	err = getRequest(req, &mr)

	if throw(w, NewError(err)) {
		return
	}

	if mr.Button == nil || *mr.Button == input.None {
		*mr.Button = input.Left
	}

	if mr.Times <= 0 {
		mr.Times = 1
	}

	var title, url, document string

	actions := make([]chromedp.Action, 0)

	actions = append(actions, network.Enable())
	var sw Watcher
	var duration time.Duration

	byPos := mr.Position != nil

	if byPos {
		var x, y float64
		x = mr.Position.X
		y = mr.Position.Y

		if mr.Selector != nil {
			var pos position
			sw = NewStopwatch()
			err = withTimeout(mr.GetTimeout(), sess,
				Position(*mr.Selector, &pos))
			duration = sw.Stop()

			if throwCapture(w, sess, NewError(errors.Wrap(err, fmt.Sprintf("element '%s' not found", *mr.Selector)))) {
				return
			}

			x += pos.Left
			y += pos.Top
		}

		switch mr.Event {
		case "click":
			actions = append(actions, chromedp.MouseClickXY(x, y, chromedp.ButtonType(*mr.Button), chromedp.ClickCount(mr.Times)))
			break
		case "mouseover":
			actions = append(actions, chromedp.Action(MouseOverXY(x, y)))
			break
		case "mouseout":
			actions = append(actions, chromedp.Action(MouseOutXY(x, y)))
			break
		case "scroll":
			actions = append(actions, chromedp.Action(chromedp.MouseEvent(input.MouseWheel, x, y)))
			break
		case "mousedown":
			actions = append(actions, chromedp.Action(chromedp.MouseEvent(input.MousePressed, x, y, chromedp.ButtonType(*mr.Button))))
			break
		case "mouseup":
			actions = append(actions, chromedp.Action(chromedp.MouseEvent(input.MouseReleased, x, y, chromedp.ButtonType(*mr.Button))))
			break
		}
	} else if mr.Selector != nil {
		var pos position
		sw = NewStopwatch()
		err = withTimeout(mr.GetTimeout(), sess,
			Position(*mr.Selector, &pos))
		duration = sw.Stop()

		if throwCapture(w, sess, NewError(errors.Wrap(err, fmt.Sprintf("element '%s' not found", *mr.Selector)))) {
			return
		}

		switch mr.Event {
		case "click":
			actions = append(actions, chromedp.MouseClickXY(pos.CenterX, pos.CenterY, chromedp.ButtonType(*mr.Button), chromedp.ClickCount(mr.Times)))
			break
		case "mouseover":
			actions = append(actions, chromedp.Action(MouseOver(*mr.Selector)))
			break
		case "mouseout":
			actions = append(actions, chromedp.Action(MouseOut(*mr.Selector)))
			break
		case "scroll":
			actions = append(actions, chromedp.Action(chromedp.ScrollIntoView(*mr.Selector, chromedp.BySearch)))
			break
		case "mousedown":
			actions = append(actions, chromedp.Action(chromedp.MouseEvent(input.MousePressed, pos.CenterX, pos.CenterY, chromedp.ButtonType(*mr.Button))))
			break
		case "mouseup":
			actions = append(actions, chromedp.Action(chromedp.MouseEvent(input.MouseReleased, pos.CenterX, pos.CenterY, chromedp.ButtonType(*mr.Button))))
			break
		}
	}

	if mr.WaitFor != nil {
		actions = append(actions, chromedp.WaitVisible(*mr.WaitFor, chromedp.BySearch))
	}

	actions = append(actions, chromedp.Title(&title),
		chromedp.Location(&url),
		chromedp.OuterHTML("html", &document))

	sw.Continue()
	err = withTimeout(mr.GetTimeout(), sess, actions...)
	duration = sw.Stop()

	response := NewResponse(duration, &title, &url, &document, nil, err)

	if throwCapture(w, sess, response) {
		return
	}

	ok(w, response)
}

type position struct {
	Dimension *dom.BoxModel
	Top       float64
	Left      float64
	Right     float64
	Bottom    float64
	CenterX   float64
	CenterY   float64
}

func (p *position) Center() {
	p.CenterX = float64(p.Dimension.Width)/2.0 + p.Left
	p.CenterY = float64(p.Dimension.Height)/2.0 + p.Top
}

func Position(selector string, pos *position) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var top, left, right, bottom string
		var found bool
		err := chromedp.Run(ctx,
			chromedp.AttributeValue(selector, "offsetLeft", &left,
				&found, chromedp.BySearch),
			chromedp.AttributeValue(selector, "offsetTop", &top,
				&found, chromedp.BySearch),
			chromedp.AttributeValue(selector, "offsetRight", &right,
				&found, chromedp.BySearch),
			chromedp.AttributeValue(selector, "offsetBottom", &bottom,
				&found, chromedp.BySearch),
			chromedp.Dimensions(selector, &pos.Dimension, chromedp.BySearch))

		if top == "" {
			top = "0"
		}
		if left == "" {
			left = "0"
		}
		if right == "" {
			right = "0"
		}
		if bottom == "" {
			bottom = "0"
		}

		var e error
		pos.Bottom, e = strconv.ParseFloat(bottom, 64)

		if err != nil && e != nil {
			err = errors.Wrap(err, e.Error())
		} else if e != nil {
			err = e
		}

		pos.Top, e = strconv.ParseFloat(top, 64)

		if err != nil && e != nil {
			err = errors.Wrap(err, e.Error())
		} else if e != nil {
			err = e
		}

		pos.Right, e = strconv.ParseFloat(right, 64)

		if err != nil && e != nil {
			err = errors.Wrap(err, e.Error())
		} else if e != nil {
			err = e
		}

		pos.Left, e = strconv.ParseFloat(left, 64)

		if err != nil && e != nil {
			err = errors.Wrap(err, e.Error())
		} else if e != nil {
			err = e
		}

		pos.Center()

		return err
	}
}

func MouseOver(selector string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var pos position
		err := chromedp.Run(ctx, Position(selector, &pos))

		if err != nil {
			return err
		}

		return chromedp.Run(ctx, MouseOverXY(pos.CenterX, pos.CenterY))
	}
}

func MouseOut(selector string) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		var pos position
		err := chromedp.Run(ctx, Position(selector, &pos))

		if err != nil {
			return err
		}

		return chromedp.Run(ctx, MouseOutXY(pos.CenterX, pos.CenterY))
	}
}

func MouseOverXY(x float64, y float64) chromedp.EvaluateAction {
	var cmd = fmt.Sprintf(`document.elementFromPoint(%f, %f).dispatchEvent(new Event('mouseover'))`,
		x, y)
	var res bool
	return chromedp.Evaluate(cmd, &res)
}

func MouseOutXY(x float64, y float64) chromedp.EvaluateAction {
	var cmd = fmt.Sprintf(`document.elementFromPoint(%f, %f).dispatchEvent(new Event('mouseout'))`,
		x, y)
	var res bool
	return chromedp.Evaluate(cmd, &res)
}
