package main

// @title Remote Chrome DevTools Protocol
// @version 1.0.0
// @host localhost:12345
// @BasePath /

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rcdp/docs"
	"sync"
	"time"

	_ "rcdp/docs"
)

var (
	sessions map[string]*session
	mut      sync.RWMutex
)

func init() {
	log.Println("initializing sessions")
	sessions = make(map[string]*session)
	mut = sync.RWMutex{}

	go watchSessions()
}

func main() {
	log.Println("initializing rcdp")
	router := mux.NewRouter()

	get := router.Methods(http.MethodGet).Subrouter()
	post := router.Methods(http.MethodPost).Subrouter()
	del := router.Methods(http.MethodDelete).Subrouter()
	swOpts := middleware.SwaggerUIOpts{
		BasePath: docs.SwaggerInfo.BasePath,
		Title:    docs.SwaggerInfo.Title,
		SpecURL:  "/swagger.yaml",
	}

	get.Handle("/docs", middleware.SwaggerUI(swOpts, nil))
	get.Handle("/swagger.yaml", http.FileServer(http.Dir("./docs")))
	get.Handle("/swagger.json", http.FileServer(http.Dir("./docs")))

	post.HandleFunc("/{session_code}/attribute", attributeEndpoint)
	del.HandleFunc("/{session_code}/close", closeEndpoint)
	post.HandleFunc("/{session_code}/inner_html", innerHtmlEndpoint)
	post.HandleFunc("/{session_code}/inner_text", innerTextEndpoint)
	post.HandleFunc("/{session_code}/javascript", javascriptEndpoint)
	post.HandleFunc("/{session_code}/mouse", mouseEndpoint)
	post.HandleFunc("/{session_code}/navigate", navigateEndpoint)
	post.HandleFunc("/{session_code}/refresh", refreshEndpoint)
	get.HandleFunc("/{session_code}/screenshot", screenshotEndpoint)
	post.HandleFunc("/{session_code}/select_option", selectOptionEndpoint)
	get.HandleFunc("/start", startEndpoint)
	post.HandleFunc("/{session_code}/type", typeEndpoint)
	post.HandleFunc("/{session_code}/viewport", viewportEndpoint)

	server := http.Server{
		Addr:    ":12345",
		Handler: router,
	}

	// Initialize the go-routine function
	go func() {
		log.Printf("Server running on port %s\n", server.Addr)
		// ListenAndServe listens on the TCP network address specified in the server property
		listenAndServeError := server.ListenAndServe()

		if listenAndServeError != nil {
			log.Fatal(listenAndServeError)
		}
	}()

	// Make the channel with type os.Signal
	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, os.Interrupt)
	signal.Notify(signalChannel, os.Kill)

	// Read the channel value
	sig := <-signalChannel

	log.Println("Received os signal, graceful timeout", sig)

	//Canceling this context releases resources associated with it
	terminateContext, terminateContextCnl := context.WithTimeout(context.Background(), 30*time.Second)
	defer terminateContextCnl()

	// Shutdown gracefully shuts down the server without interrupting any active connections
	err := server.Shutdown(terminateContext)

	log.Fatal(err)
}
