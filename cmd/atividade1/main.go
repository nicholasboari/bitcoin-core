package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// aqui retorna o resumo da mempool
	http.HandleFunc("/api/mempool/summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cmd := bitcoinCLICommand("getrawmempool", "true")

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

		cmd := bitcoinCLICommand("getblockchaininfo")

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

	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}

	http.Handle("/", http.FileServer(http.FS(staticRoot)))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func bitcoinCLICommand(args ...string) *exec.Cmd {
	commandArgs := append(bitcoinCLIBaseArgs(), args...)
	return exec.Command("bitcoin-cli", commandArgs...)
}

func bitcoinCLIBaseArgs() []string {
	args := bitcoinCLINetworkArgs()

	if value := os.Getenv("BITCOIN_RPC_CONNECT"); value != "" {
		args = append(args, "-rpcconnect="+value)
	}
	if value := os.Getenv("BITCOIN_RPC_PORT"); value != "" {
		args = append(args, "-rpcport="+value)
	}
	if value := os.Getenv("BITCOIN_RPC_USER"); value != "" {
		args = append(args, "-rpcuser="+value)
	}
	if value := os.Getenv("BITCOIN_RPC_PASSWORD"); value != "" {
		args = append(args, "-rpcpassword="+value)
	}

	return args
}

func bitcoinCLINetworkArgs() []string {
	network := os.Getenv("BITCOIN_NETWORK")
	if network == "" {
		network = "regtest"
	}
	if network == "mainnet" {
		return nil
	}
	return []string{"-" + network}
}
