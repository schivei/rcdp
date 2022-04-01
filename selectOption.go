package main

import (
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type option struct {
	position int
	value    string
	text     string
	selected bool
}

func selectOptions(options *map[int]*option) func(i int, selection *goquery.Selection) {
	return func(i int, selection *goquery.Selection) {
		value, exists := selection.Attr("value")
		text := selection.RemoveAttr("selected").Text()

		if !exists {
			value = strings.TrimSpace(text)
		}

		(*options)[i] = &option{position: i, value: value, text: text, selected: false}
	}
}

func setOptionSelection(options *map[int]*option, c string,
	validate func(o *option, c string) bool) {
	for _, o := range *options {
		o.selected = o.selected || validate(o, c)
	}
}

// selectOptionEndpoint ... Select
// @Summary Select
// @Description Select
// @Tags Select
// @Success 200 {object} dataResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /{session_code}/select_option [post] {object} selectRequest
// @Body {object} selectRequest
// @Param content body selectRequest true "content data"
func selectOptionEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("selecting options")
	if !allowed(w, req, http.MethodPost) {
		return
	}
	var selectBy string

	sess, err := getSession(req)
	if throw(w, NewError(err)) {
		return
	}

	var sr selectRequest
	err = getRequest(req, &sr)

	if throw(w, NewError(err)) {
		return
	}

	var title, url, document string

	actions := make([]chromedp.Action, 0)

	actions = append(actions, network.Enable())
	var sw Watcher
	var duration time.Duration

	sw = NewStopwatch()
	err = withTimeout(sr.GetTimeout(), sess,
		chromedp.OuterHTML(sr.Selector, &document, chromedp.BySearch))
	duration = sw.Stop()

	if throwCapture(w, sess, NewError(err)) {
		return
	}

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(document))

	if throwCapture(w, sess, NewError(err)) {
		return
	}

	options := make(map[int]*option)

	opts := doc.Find("option")

	opts.Each(selectOptions(&options))

	if len(options) > 0 {
		ops := sr.Data

		switch {
		case selectBy == "index":
			for _, c := range ops {
				setOptionSelection(&options, c, func(o *option, c string) bool {
					index, e := strconv.Atoi(c)

					return e == nil && o.position == index
				})
			}
			break
		case selectBy == "text":
			for _, c := range ops {
				setOptionSelection(&options, c, func(o *option, c string) bool {
					return o.text == c
				})
			}
			break
		case selectBy == "partialtext":
			for _, c := range ops {
				setOptionSelection(&options, c, func(o *option, c string) bool {
					return strings.Contains(o.text, c)
				})
			}
			break
		case selectBy == "regex":
			for _, c := range ops {
				setOptionSelection(&options, c, func(o *option, c string) bool {
					reg, e := regexp.Compile(c)
					return e == nil && (reg.MatchString(o.text) || reg.MatchString(o.value))
				})
			}
			break
		case selectBy == "value":
			for _, c := range ops {
				setOptionSelection(&options, c, func(o *option, c string) bool {
					return o.value == c
				})
			}
			break
		}

		slb, er1 := json.Marshal(sr.Selector)

		if throw(w, NewError(er1)) {
			return
		}

		sel := "$(document).xpathEvaluate(" + string(slb) + ").val("

		selections := make([]string, 0)

		for _, o := range options {
			if o.selected {
				selections = append(selections, o.value)
			}
		}

		vlb, er2 := json.Marshal(selections)

		if throw(w, NewError(er2)) {
			return
		}

		sel += string(vlb) + ");"

		var res interface{}

		if sr.WaitFor != nil {
			actions = append(actions, chromedp.WaitVisible(*sr.WaitFor, chromedp.BySearch))
		}

		actions = append(actions, chromedp.Title(&title),
			chromedp.Location(&url),
			chromedp.Evaluate(`function lumini360_collector_select(){
$.fn.xpathEvaluate = function (xpathExpression) {
   // NOTE: vars not declared local for debug purposes
   $this = this.first(); // Don't make me deal with multiples before coffee

   // Evaluate xpath and retrieve matching nodes
   xpathResult = this[0].evaluate(xpathExpression, this[0], null, XPathResult.ORDERED_NODE_ITERATOR_TYPE, null);

   result = [];
   while (elem = xpathResult.iterateNext()) {
      result.push(elem);
   }

   $result = jQuery([]).pushStack( result );
   return $result;
};
`+sel+`
return 0;
}
lumini360_collector_select()`, &res, chromedp.EvalAsValue))

		sw.Continue()
		err = withTimeout(sr.GetTimeout(), sess, actions...)
		duration = sw.Stop()

		response := NewResponse(duration, &title, &url, &document, nil, err)

		if throwCapture(w, sess, response) {
			return
		}

		ok(w, response)
	}

	throwCapture(w, sess, NewError(errors.New("select option not found")))
}
