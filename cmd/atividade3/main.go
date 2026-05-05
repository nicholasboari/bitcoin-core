package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

const httpAddress = ":8082"

func main() {
	http.HandleFunc("/wallets", walletsHandler)
	http.HandleFunc("/wallet/select", selectWalletHandler)

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
