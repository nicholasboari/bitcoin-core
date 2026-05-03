package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

func main() {
	// aqui retorna o resumo da mempool
	http.HandleFunc("/api/mempool/summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cmd := exec.Command("bitcoin-cli", "-regtest", "getrawmempool", "true")

		output, err := cmd.CombinedOutput()
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("erro ao executar bitcoin-cli: %v\nsaida: %s", err, string(output)),
				http.StatusInternalServerError,
			)
			return
		}

		var mempool map[string]MempoolTx

		if err := json.Unmarshal(output, &mempool); err != nil {
			http.Error(w, fmt.Sprintf("erro ao parsear JSON: %v", err), http.StatusInternalServerError)
			return
		}

		summary := calculateMempoolSummary(mempool)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	})

	// aqui retorna n° de blocos, n° de headers e diferença entre eles
	http.HandleFunc("/api/blockchain/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cmd := exec.Command("bitcoin-cli", "-regtest", "getblockchaininfo")

		output, err := cmd.CombinedOutput()
		if err != nil {
			http.Error(
				w,
				fmt.Sprintf("erro ao executar bitcoin-cli: %v\nsaida: %s", err, string(output)),
				http.StatusInternalServerError,
			)
			return
		}

		var blockchainInfo BlockchainInfo

		if err := json.Unmarshal(output, &blockchainInfo); err != nil {
			http.Error(w, fmt.Sprintf("erro ao parsear JSON: %v", err), http.StatusInternalServerError)
			return
		}

		blockchainInfo.BlocksToHeaders = blockchainInfo.Headers - blockchainInfo.Blocks

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(blockchainInfo)
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
