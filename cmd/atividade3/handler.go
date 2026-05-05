package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sync"
)

var (
	activeWallet   string
	activeWalletMu sync.RWMutex
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

func selectWalletHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request SelectWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "json invalido", http.StatusBadRequest)
		return
	}

	if request.Wallet == "" {
		http.Error(w, "wallet e obrigatoria", http.StatusBadRequest)
		return
	}

	availableWallets, err := getAvailableWallets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !contains(availableWallets, request.Wallet) {
		http.Error(w, "wallet nao encontrada", http.StatusNotFound)
		return
	}

	loadedWallets, err := getLoadedWallets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !contains(loadedWallets, request.Wallet) {
		if err := loadWallet(request.Wallet); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	setActiveWallet(request.Wallet)

	walletInfo, err := getWalletInfo(request.Wallet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := SelectWalletResponse{
		SelectedWallet: request.Wallet,
		WalletInfo:     walletInfo,
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

func loadWallet(wallet string) error {
	cmd := exec.Command("bitcoin-cli", "-regtest", "loadwallet", wallet)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erro ao executar bitcoin-cli loadwallet: %v\nsaida: %s", err, string(output))
	}

	return nil
}

func getWalletInfo(wallet string) (WalletInfo, error) {
	cmd := exec.Command("bitcoin-cli", "-regtest", "-rpcwallet="+wallet, "getwalletinfo")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return WalletInfo{}, fmt.Errorf("erro ao executar bitcoin-cli getwalletinfo: %v\nsaida: %s", err, string(output))
	}

	var walletInfo WalletInfo
	if err := json.Unmarshal(output, &walletInfo); err != nil {
		return WalletInfo{}, fmt.Errorf("erro ao parsear getwalletinfo: %v", err)
	}

	return walletInfo, nil
}

func selectedWallet(availableWallets []string, loadedWallets []string) string {
	activeWalletMu.RLock()
	defer activeWalletMu.RUnlock()

	if activeWallet != "" && contains(availableWallets, activeWallet) {
		return activeWallet
	}

	if len(loadedWallets) > 0 {
		return loadedWallets[0]
	}

	if len(availableWallets) > 0 {
		return availableWallets[0]
	}

	return ""
}

func setActiveWallet(wallet string) {
	activeWalletMu.Lock()
	defer activeWalletMu.Unlock()

	activeWallet = wallet
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}

	return false
}
