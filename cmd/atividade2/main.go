package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

const (
	blockAddress = "tcp://127.0.0.1:28332"
	txAddress    = "tcp://127.0.0.1:28333"

	blockTopic = "hashblock"
	txTopic    = "hashtx"

	httpAddress     = ":8081"
	maxStoredEvents = 10_000
)

func main() {
	ctx := context.Background()
	store := NewEventStore(maxStoredEvents)

	// aqui inicia 2 goroutines para escutar os tópicos hashblock e hashtx
	go listen(ctx, blockAddress, blockTopic, store)
	go listen(ctx, txAddress, txTopic, store)

	http.HandleFunc("/api/events/summary", eventsSummaryHandler(store))
	http.HandleFunc("/api/events/latest", latestEventsHandler(store))
	http.HandleFunc("/api/events/state-comparison", stateComparisonHandler(store))

	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	http.Handle("/", http.FileServer(http.FS(staticRoot)))

	fmt.Println("http listening on", httpAddress)
	if err := http.ListenAndServe(httpAddress, nil); err != nil {
		panic(err)
	}
}
