package main

import (
	"context"
	"go_event_loop/index"
	"go_event_loop/runtime"
	"net/http"
)

func main() {
	runtime.Init()
	go runtime.RunEventLoop(context.Background())
	var mux http.ServeMux
	mux.Handle("/{$}", index.Handler())
	server := http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: &mux,
	}
	server.ListenAndServe()
}
