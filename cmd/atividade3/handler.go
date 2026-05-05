package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
)

func walletsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	availableWallets, err := getAvailableWallets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	loadedWallets, err := getLoadedWallets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := WalletsResponse{
		AvailableWallets: availableWallets,
		LoadedWallets:    loadedWallets,
		SelectedWallet:   selectedWallet(availableWallets, loadedWallets),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getAvailableWallets() ([]string, error) {
	cmd := exec.Command("bitcoin-cli", "-regtest", "listwalletdir")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("erro ao executar bitcoin-cli listwalletdir: %v\nsaida: %s", err, string(output))
	}

	var result ListWalletDirResponse
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("erro ao parsear listwalletdir: %v", err)
	}

	wallets := make([]string, 0, len(result.Wallets))
	for _, wallet := range result.Wallets {
		wallets = append(wallets, wallet.Name)
	}

	return wallets, nil
}

func getLoadedWallets() ([]string, error) {
	cmd := exec.Command("bitcoin-cli", "-regtest", "listwallets")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("erro ao executar bitcoin-cli listwallets: %v\nsaida: %s", err, string(output))
	}

	var wallets []string
	if err := json.Unmarshal(output, &wallets); err != nil {
		return nil, fmt.Errorf("erro ao parsear listwallets: %v", err)
	}

	return wallets, nil
}

func selectedWallet(availableWallets []string, loadedWallets []string) string {
	if len(loadedWallets) > 0 {
		return loadedWallets[0]
	}

	if len(availableWallets) > 0 {
		return availableWallets[0]
	}

	return ""
}
