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
	if err := loadTrackedTransactions(); err != nil {
		panic(err)
	}

	http.HandleFunc("/wallets", walletsHandler)
	http.HandleFunc("/wallet/select", selectWalletHandler)
	http.HandleFunc("/wallet/status", walletStatusHandler)
	http.HandleFunc("/tx/send", sendTransactionHandler)
	http.HandleFunc("/txs", transactionsHandler)
	http.HandleFunc("/tx/", transactionHandler)

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
