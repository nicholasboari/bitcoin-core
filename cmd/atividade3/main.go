package main

import (
	"fmt"
	"net/http"
)

const httpAddress = ":8082"

func main() {
	http.HandleFunc("/wallets", walletsHandler)

	fmt.Println("http listening on", httpAddress)
	if err := http.ListenAndServe(httpAddress, nil); err != nil {
		panic(err)
	}
}
