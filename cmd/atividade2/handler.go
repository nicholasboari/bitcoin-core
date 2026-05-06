package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const defaultSummarySeconds = 60
const defaultLatestEvents = 10

func eventsSummaryHandler(store *EventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := r.URL.Query()

		seconds := defaultSummarySeconds
		if rawSeconds := query.Get("seconds"); rawSeconds != "" {
			parsedSeconds, ok := parsePositiveInt(rawSeconds)
			if !ok {
				http.Error(w, "seconds must be a positive integer", http.StatusBadRequest)
				return
			}

			seconds = parsedSeconds
		}

		summary := store.SummaryLastSeconds(seconds, time.Now())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}

func latestEventsHandler(store *EventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		latest := store.Latest(defaultLatestEvents)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(latest)
	}
}

func stateComparisonHandler(store *EventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		bestBlock, err := getBestBlockHash()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		lastSeenBlock := store.LastSeenBlock()
		comparison := StateComparison{
			BestBlock:     bestBlock,
			LastSeenBlock: lastSeenBlock,
			Divergence:    bestBlock != lastSeenBlock,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comparison)
	}
}

func parsePositiveInt(value string) (int, bool) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, false
	}

	return parsed, true
}

func getBestBlockHash() (string, error) {
	commandArgs := append(bitcoinCLIBaseArgs(), "getbestblockhash")
	cmd := exec.Command("bitcoin-cli", commandArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("erro ao executar bitcoin-cli: %v\nsaida: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
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
