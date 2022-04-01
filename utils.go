package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type internalError struct {
	Message    string         `json:"message"`
	Cause      *string        `json:"cause,omitempty"`
	InnerError *internalError `json:"inner_error,omitempty"`
}

type dataResponse struct {
	Error      *internalError `json:"error"`
	Time       *int64         `json:"time,omitempty"`
	Title      *string        `json:"title,omitempty"`
	Url        *string        `json:"url,omitempty"`
	Document   *string        `json:"document,omitempty"`
	Evidence   *string        `json:"evidence,omitempty"`
	LocalError error          `json:"-"`
}

type IDataResponse interface {
	HasError() bool
	GetError() error
	AppendError(err error)
	SetEvidence(evidence *string)
	AppendTime(duration int64)
	SerializeErrors()
}

func writeResponse(w http.ResponseWriter, statusCode int, dr IDataResponse) {
	w.Header().Del("Content-Type")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	dr.SerializeErrors()

	if dr.HasError() {
		log.Println(dr.GetError())
	}

	err := json.NewEncoder(w).Encode(dr)
	if err != nil {
		log.Println(err)
	}
}

func (dr *dataResponse) GetError() error {
	return dr.LocalError
}

func (dr *dataResponse) HasError() bool {
	return dr.LocalError != nil
}

func (dr *dataResponse) AppendError(err error) {
	if err != nil {
		dr.LocalError = errors.Wrap(dr.LocalError, err.Error())
	}
}

func (dr *dataResponse) SetEvidence(evidence *string) {
	dr.Evidence = evidence
}

func (dr *dataResponse) AppendTime(duration int64) {
	var tm int64
	if dr.Time != nil {
		tm = *dr.Time
	}
	eta := tm + duration
	dr.Time = &eta
}

// Data Request
type dataRequest struct {
	// Timeout in nanoseconds
	Timeout int64 `json:"timeout,omitempty" example:"1000000000" format:"int64"`
}

type IDataRequest interface {
	GetTimeout() int64
}

func (dr *dataRequest) GetTimeout() int64 {
	return dr.Timeout
}

func throwCapture(w http.ResponseWriter, sess *session, response IDataResponse) bool {
	if response.HasError() {
		img, eta, e := takeScreenshot(sess, nil)
		if e != nil {
			response.AppendError(e)
		}
		response.AppendTime(eta.Nanoseconds())
		response.SetEvidence(img)
		badExecution(w, response)
		return true
	}
	return false
}

func throw(w http.ResponseWriter, response IDataResponse) bool {
	if response.HasError() {
		badRequest(w, response)
		return true
	}
	return false
}

func getRequest(req *http.Request, data IDataRequest) error {
	return json.NewDecoder(req.Body).Decode(data)
}

func GetError(err error) *internalError {
	if err == nil {
		return nil
	}

	causeErr := errors.Cause(err)
	var cause *string
	if causeErr != nil && causeErr.Error() != err.Error() {
		c := causeErr.Error()
		cause = &c
	}

	ie := &internalError{
		Cause:      cause,
		Message:    err.Error(),
		InnerError: GetError(errors.Unwrap(err)),
	}

	return ie
}

func NewErrorTime(err error, duration time.Duration) IDataResponse {
	r := NewError(err)
	r.AppendTime(duration.Nanoseconds())
	return r
}

func NewError(err error) IDataResponse {
	return &dataResponse{
		Time:       nil,
		Title:      nil,
		Document:   nil,
		Evidence:   nil,
		Url:        nil,
		LocalError: err,
	}
}

func (dr *dataResponse) SerializeErrors() {
	dr.Error = GetError(dr.LocalError)
}

func NewResponse(duration time.Duration, title, url, document, evidence *string, err error) IDataResponse {
	ns := duration.Nanoseconds()
	var eta *int64 = nil
	if ns > 0 {
		eta = &ns
	}

	if title != nil && len(strings.TrimSpace(*title)) == 0 {
		title = nil
	}

	if url != nil && len(strings.TrimSpace(*url)) == 0 {
		url = nil
	}

	if document != nil && len(strings.TrimSpace(*document)) == 0 {
		document = nil
	}

	if evidence != nil && len(strings.TrimSpace(*evidence)) == 0 {
		evidence = nil
	}

	return &dataResponse{
		Time:       eta,
		Title:      title,
		Document:   document,
		Evidence:   evidence,
		Url:        url,
		LocalError: err,
	}
}

func ok(w http.ResponseWriter, response IDataResponse) {
	writeResponse(w, 200, response)
}

func badRequest(w http.ResponseWriter, response IDataResponse) {
	writeResponse(w, 400, response)
}

func badExecution(w http.ResponseWriter, response IDataResponse) {
	writeResponse(w, 500, response)
}

func allowed(w http.ResponseWriter, req *http.Request, method string) bool {
	if req.Method != method {
		return !throw(w, NewError(errors.New("method not allowed")))
	}
	return true
}

func getSession(req *http.Request) (*session, error) {
	params := mux.Vars(req)
	sessionCode := params["session_code"]

	mut.RLock()
	defer mut.RUnlock()
	sess, found := sessions[sessionCode]
	if !found {
		return nil, errors.New("session closed")
	}

	if (runtime.GOOS == "windows" && (sess.acnl == nil || sess.actx == nil)) ||
		sess.ccnl == nil || sess.cctx == nil {
		return nil, errors.New("session closed")
	}

	sess.last = time.Now()

	return sess, nil
}

func withTimeout(timeout int64, sess *session, actions ...chromedp.Action) error {
	if timeout <= 0 {
		return chromedp.Run(sess.cctx, actions...)
	}

	tctx, cancel := context.WithTimeout(sess.cctx, time.Duration(timeout)*time.Nanosecond)
	defer cancel()

	done := make(chan error)

	go func(ctx context.Context, actions []chromedp.Action) {
		done <- chromedp.Run(ctx, actions...)
	}(tctx, actions)

	for {
		select {
		case <-tctx.Done():
			return tctx.Err()
		case err := <-done:
			return err
		}
	}
}

func takeScreenshot(sess *session, selector *string) (*string, time.Duration, error) {
	var err error
	pic := make([]byte, 0)
	var eta time.Duration
	if selector == nil || len(strings.TrimSpace(*selector)) == 0 {
		sw := NewStopwatch()
		err = chromedp.Run(sess.cctx, chromedp.FullScreenshot(&pic, 100))
		eta = sw.Stop()
	} else {
		sw := NewStopwatch()
		err = chromedp.Run(sess.cctx, chromedp.Screenshot(*selector, &pic, chromedp.BySearch))
		eta = sw.Stop()
	}

	return toBase64(pic), eta, err
}

func toBase64(image []byte) *string {
	if image != nil && len(image) > 0 {
		img := base64.StdEncoding.EncodeToString(image)
		return &img
	}

	return nil
}

func deleteSession(req *http.Request) {
	params := mux.Vars(req)
	sessionCode := params["session_code"]
	go dropSession(sessionCode)
}

func recovery() {
	if a := recover(); a != nil {
		fmt.Println("RECOVER", a)
	}
}

func collect() {
	defer recovery()
	runtime.GC()
}

func dropSession(sessionCode string) {
	defer recovery()
	defer collect()
	log.Println("dropping session")

	mut.Lock()
	defer mut.Unlock()
	sess, found := sessions[sessionCode]

	if !found {
		return
	}

	if sess == nil {
		return
	}

	if sess.ccnl != nil {
		sess.ccnl()
		sess.ccnl = nil
	}

	if sess.acnl != nil {
		sess.acnl()
		sess.acnl = nil
	}

	userDataDir := sess.usdd

	sess = nil

	delete(sessions, sessionCode)

	go dropFolder(userDataDir)
}

func dropFolder(userDataDir string) {
	time.Sleep(1 * time.Minute)
	log.Println("dropping folder")
	_ = os.RemoveAll(userDataDir)
}

func watchSessions() {
	log.Println("watching expired sessions")
	for {
		time.Sleep(1 * time.Second)
		mut.Lock()
		for sessionCode, sess := range sessions {
			last := sess.last
			sess = nil
			if time.Now().Sub(last).Hours() >= 1 {
				go dropSession(sessionCode)
			}
		}
		mut.Unlock()
	}
}

type Watcher interface {
	Stop() time.Duration
	Restart()
	Continue()
}

type watch struct {
	time      time.Time
	stopped   *time.Time
	timeStops time.Duration
}

func NewStopwatch() Watcher {
	return &watch{time: time.Now(), stopped: nil}
}

func (w *watch) Stop() time.Duration {
	now := time.Now()
	w.stopped = &now
	return now.Sub(w.time) - w.timeStops
}

func (w *watch) Restart() {
	w.stopped = nil
	w.timeStops = 0
	w.time = time.Now()
}

func (w *watch) Continue() {
	if w.stopped != nil {
		now := time.Now()
		w.timeStops += now.Sub(*w.stopped)
	}
	w.stopped = nil
}
