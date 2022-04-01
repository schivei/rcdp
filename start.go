package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// startEndpoint ... Start Browser
// @Summary Start Browser
// @Description Start Browser
// @Tags Start
// @Success 200 {object} startResponse
// @Failure 400 {object} dataResponse
// @Failure 500 {object} dataResponse
// @Router /start [get]
func startEndpoint(w http.ResponseWriter, req *http.Request) {
	log.Println("starting session")
	if !allowed(w, req, http.MethodGet) {
		return
	}

	sessionCode := uuid.New().String()

	response := &startResponse{
		SessionCode:  sessionCode,
		dataResponse: &dataResponse{},
	}

	userDataDir, err := ioutil.TempDir("", sessionCode)

	response.AppendError(err)

	if throw(w, response) {
		return
	}

	options := make([]chromedp.ExecAllocatorOption, 0)
	options = append(options, chromedp.UserDataDir(userDataDir),
		chromedp.NoSandbox,
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.DisableGPU,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.IgnoreCertErrors)

	if runtime.GOOS == "windows" {
		edge := "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe"

		if found, e := exec.LookPath(edge); e == nil {
			options = append(options, chromedp.ExecPath(found))
		}
	} else {
		options = append(options, chromedp.Headless)
	}

	var actx context.Context
	var acnl context.CancelFunc

	headless := "/headless-shell/headless-shell"

	if _, e := exec.LookPath(headless); e == nil {
		actx, acnl = chromedp.NewRemoteAllocator(context.Background(), "ws://127.0.0.1:9222")
		go dropFolder(userDataDir)
	} else {
		actx, acnl = chromedp.NewExecAllocator(
			context.Background(),
			options...,
		)
	}

	cctx, ccnl := chromedp.NewContext(actx, chromedp.WithLogf(log.Printf))

	sess := &session{
		usdd: userDataDir,
		acnl: acnl,
		actx: actx,
		ccnl: ccnl,
		cctx: cctx,
		date: time.Now(),
		last: time.Now(),
	}

	mut.Lock()
	defer mut.Unlock()
	sessions[sessionCode] = sess

	sw := NewStopwatch()
	var url, title, document string
	err = chromedp.Run(sess.cctx,
		chromedp.Navigate("about:blank"),
		chromedp.Location(&url),
		chromedp.Title(&title),
		chromedp.OuterHTML("html", &document))
	duration := sw.Stop()

	resp := NewResponse(duration, &title, &url, &document, nil, err)

	response.dataResponse, _ = resp.(*dataResponse)

	if throw(w, response) {
		go dropSession(sessionCode)
		return
	}

	ok(w, response)
}
