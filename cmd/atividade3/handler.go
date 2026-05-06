package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	defaultFeeSats          = int64(1000)
	trackedTransactionsFile = "tracked_transactions.json"
)

var (
	activeWallet    string
	activeWalletMu  sync.RWMutex
	trackedTxs      = make(map[string]TrackedTransaction)
	trackedTxsOrder []string
	trackedTxsMu    sync.RWMutex
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

	if response.SelectedWallet != "" {
		setActiveWallet(response.SelectedWallet)
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

func walletStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wallet, err := ensureActiveWallet()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	walletInfo, err := getWalletInfo(wallet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utxos, err := listUnspent(wallet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := WalletStatusResponse{
		Wallet:  wallet,
		Balance: walletInfo.Balance,
		UTXOs:   len(utxos),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request SendTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "json invalido", http.StatusBadRequest)
		return
	}

	request.Address = strings.TrimSpace(request.Address)
	if request.Address == "" {
		http.Error(w, "address e obrigatorio", http.StatusBadRequest)
		return
	}

	amountSats := btcToSats(request.Amount)
	if amountSats <= 0 {
		http.Error(w, "amount deve ser maior que zero", http.StatusBadRequest)
		return
	}

	wallet, err := ensureActiveWallet()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rawTx, err := createFundedRawTransaction(wallet, request.Address, amountSats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	signedTx, err := signRawTransaction(wallet, rawTx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	txid, err := sendRawTransaction(signedTx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := trackTransaction(txid, wallet); err != nil {
		http.Error(w, fmt.Sprintf("transacao enviada (%s), mas erro ao persistir historico: %v", txid, err), http.StatusInternalServerError)
		return
	}

	response := SendTransactionResponse{
		TxID:   txid,
		Wallet: wallet,
		RawTx:  signedTx,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func transactionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tracked := allTrackedTransactions()
	transactions := make([]TransactionResponse, 0, len(tracked))
	for _, tx := range tracked {
		transactions = append(transactions, buildTransactionResponse(tx.TxID, tx.Wallet))
	}

	response := TransactionsResponse{Transactions: transactions}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func transactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	txid := strings.TrimPrefix(r.URL.Path, "/tx/")
	if txid == "" || txid == r.URL.Path {
		http.Error(w, "txid e obrigatorio", http.StatusBadRequest)
		return
	}

	wallet := getTrackedWallet(txid)
	if wallet == "" {
		var err error
		wallet, err = ensureActiveWallet()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	response := buildTransactionResponse(txid, wallet)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getAvailableWallets() ([]string, error) {
	output, err := runNodeRPC("listwalletdir")
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
	output, err := runNodeRPC("listwallets")
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
	output, err := runNodeRPC("loadwallet", wallet)
	if err != nil {
		return fmt.Errorf("erro ao executar bitcoin-cli loadwallet: %v\nsaida: %s", err, string(output))
	}

	return nil
}

func getWalletInfo(wallet string) (WalletInfo, error) {
	output, err := runWalletRPC(wallet, "getwalletinfo")
	if err != nil {
		return WalletInfo{}, fmt.Errorf("erro ao executar bitcoin-cli getwalletinfo: %v\nsaida: %s", err, string(output))
	}

	var walletInfo WalletInfo
	if err := json.Unmarshal(output, &walletInfo); err != nil {
		return WalletInfo{}, fmt.Errorf("erro ao parsear getwalletinfo: %v", err)
	}

	return walletInfo, nil
}

func listUnspent(wallet string) ([]UTXO, error) {
	output, err := runWalletRPC(wallet, "listunspent")
	if err != nil {
		return nil, fmt.Errorf("erro ao executar bitcoin-cli listunspent: %v\nsaida: %s", err, string(output))
	}

	var utxos []UTXO
	if err := json.Unmarshal(output, &utxos); err != nil {
		return nil, fmt.Errorf("erro ao parsear listunspent: %v", err)
	}

	return utxos, nil
}

func getRawChangeAddress(wallet string) (string, error) {
	output, err := runWalletRPC(wallet, "getrawchangeaddress")
	if err != nil {
		return "", fmt.Errorf("erro ao executar bitcoin-cli getrawchangeaddress: %v\nsaida: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func createFundedRawTransaction(wallet string, address string, amountSats int64) (string, error) {
	utxos, err := listUnspent(wallet)
	if err != nil {
		return "", err
	}

	targetSats := amountSats + defaultFeeSats
	selectedInputs := make([]RawTxInput, 0)
	selectedSats := int64(0)

	for _, utxo := range utxos {
		selectedInputs = append(selectedInputs, RawTxInput{TxID: utxo.TxID, Vout: utxo.Vout})
		selectedSats += btcToSats(utxo.Amount)
		if selectedSats >= targetSats {
			break
		}
	}

	if selectedSats < targetSats {
		return "", fmt.Errorf("saldo insuficiente: necessario %.8f BTC incluindo fee fixa", satsToBTC(targetSats))
	}

	outputs := map[string]float64{
		address: satsToBTC(amountSats),
	}

	changeSats := selectedSats - amountSats - defaultFeeSats
	if changeSats > 0 {
		changeAddress, err := getRawChangeAddress(wallet)
		if err != nil {
			return "", err
		}
		outputs[changeAddress] = satsToBTC(changeSats)
	}

	inputsJSON, err := json.Marshal(selectedInputs)
	if err != nil {
		return "", fmt.Errorf("erro ao montar inputs: %v", err)
	}

	outputsJSON, err := json.Marshal(outputs)
	if err != nil {
		return "", fmt.Errorf("erro ao montar outputs: %v", err)
	}

	output, err := runNodeRPC("createrawtransaction", string(inputsJSON), string(outputsJSON))
	if err != nil {
		return "", fmt.Errorf("erro ao executar bitcoin-cli createrawtransaction: %v\nsaida: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func signRawTransaction(wallet string, rawTx string) (string, error) {
	output, err := runWalletRPC(wallet, "signrawtransactionwithwallet", rawTx)
	if err != nil {
		return "", fmt.Errorf("erro ao executar bitcoin-cli signrawtransactionwithwallet: %v\nsaida: %s", err, string(output))
	}

	var result SignRawTransactionResponse
	if err := json.Unmarshal(output, &result); err != nil {
		return "", fmt.Errorf("erro ao parsear signrawtransactionwithwallet: %v", err)
	}

	if !result.Complete {
		return "", fmt.Errorf("assinatura incompleta para a wallet %s", wallet)
	}

	return result.Hex, nil
}

func sendRawTransaction(rawTx string) (string, error) {
	output, err := runNodeRPC("sendrawtransaction", rawTx)
	if err != nil {
		return "", fmt.Errorf("erro ao executar bitcoin-cli sendrawtransaction: %v\nsaida: %s", err, string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func getWalletTransaction(wallet string, txid string) (WalletTransaction, error) {
	output, err := runWalletRPC(wallet, "gettransaction", txid)
	if err != nil {
		return WalletTransaction{}, err
	}

	var transaction WalletTransaction
	if err := json.Unmarshal(output, &transaction); err != nil {
		return WalletTransaction{}, fmt.Errorf("erro ao parsear gettransaction: %v", err)
	}

	return transaction, nil
}

func getMempoolEntry(txid string) (MempoolEntry, error) {
	output, err := runNodeRPC("getmempoolentry", txid)
	if err != nil {
		return MempoolEntry{}, err
	}

	var entry MempoolEntry
	if err := json.Unmarshal(output, &entry); err != nil {
		return MempoolEntry{}, fmt.Errorf("erro ao parsear getmempoolentry: %v", err)
	}

	return entry, nil
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

func ensureActiveWallet() (string, error) {
	activeWalletMu.RLock()
	wallet := activeWallet
	activeWalletMu.RUnlock()

	availableWallets, err := getAvailableWallets()
	if err != nil {
		return "", err
	}

	loadedWallets, err := getLoadedWallets()
	if err != nil {
		return "", err
	}

	if wallet == "" || !contains(availableWallets, wallet) {
		wallet = selectedWallet(availableWallets, loadedWallets)
	}

	if wallet == "" {
		return "", fmt.Errorf("nenhuma wallet disponivel")
	}

	if !contains(availableWallets, wallet) {
		return "", fmt.Errorf("wallet selecionada nao encontrada")
	}

	if !contains(loadedWallets, wallet) {
		if err := loadWallet(wallet); err != nil {
			return "", err
		}
	}

	setActiveWallet(wallet)
	return wallet, nil
}

func setActiveWallet(wallet string) {
	activeWalletMu.Lock()
	defer activeWalletMu.Unlock()

	activeWallet = wallet
}

func loadTrackedTransactions() error {
	content, err := os.ReadFile(trackedTransactionsFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	var transactions []TrackedTransaction
	if err := json.Unmarshal(content, &transactions); err != nil {
		return fmt.Errorf("erro ao parsear %s: %v", trackedTransactionsFile, err)
	}

	trackedTxsMu.Lock()
	defer trackedTxsMu.Unlock()

	trackedTxs = make(map[string]TrackedTransaction, len(transactions))
	trackedTxsOrder = make([]string, 0, len(transactions))
	for _, transaction := range transactions {
		if transaction.TxID == "" || transaction.Wallet == "" {
			continue
		}
		if _, exists := trackedTxs[transaction.TxID]; exists {
			continue
		}

		trackedTxs[transaction.TxID] = transaction
		trackedTxsOrder = append(trackedTxsOrder, transaction.TxID)
	}

	return nil
}

func trackTransaction(txid string, wallet string) error {
	trackedTxsMu.Lock()
	defer trackedTxsMu.Unlock()

	if _, exists := trackedTxs[txid]; !exists {
		trackedTxsOrder = append([]string{txid}, trackedTxsOrder...)
	}

	trackedTxs[txid] = TrackedTransaction{
		TxID:   txid,
		Wallet: wallet,
		SentAt: time.Now().Unix(),
	}

	transactions := trackedTransactionsSnapshotLocked()
	return saveTrackedTransactions(transactions)
}

func saveTrackedTransactions(transactions []TrackedTransaction) error {
	content, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar historico de transacoes: %v", err)
	}

	return os.WriteFile(trackedTransactionsFile, content, 0o600)
}

func trackedTransactionsSnapshotLocked() []TrackedTransaction {
	transactions := make([]TrackedTransaction, 0, len(trackedTxsOrder))
	for _, txid := range trackedTxsOrder {
		transactions = append(transactions, trackedTxs[txid])
	}

	return transactions
}

func getTrackedWallet(txid string) string {
	trackedTxsMu.RLock()
	defer trackedTxsMu.RUnlock()

	return trackedTxs[txid].Wallet
}

func allTrackedTransactions() []TrackedTransaction {
	trackedTxsMu.RLock()
	defer trackedTxsMu.RUnlock()

	transactions := make([]TrackedTransaction, 0, len(trackedTxsOrder))
	for _, txid := range trackedTxsOrder {
		transactions = append(transactions, trackedTxs[txid])
	}

	return transactions
}

func buildTransactionResponse(txid string, wallet string) TransactionResponse {
	now := time.Now().Unix()
	response := TransactionResponse{
		TxID:   txid,
		Wallet: wallet,
		Status: "unknown",
	}

	if transaction, err := getWalletTransaction(wallet, txid); err == nil {
		response.Confirmations = transaction.Confirmations
		response.BlockHash = transaction.BlockHash
		response.Confirmed = transaction.Confirmations > 0
		response.AgeSeconds = ageSeconds(txid, transaction.Time, now)

		if response.Confirmed {
			response.Status = "confirmed"
			response.Message = "Transação confirmada em bloco."
			return response
		}
	}

	if entry, err := getMempoolEntry(txid); err == nil {
		response.Status = "mempool"
		response.Confirmed = false
		response.Confirmations = 0
		response.BlockHash = ""
		response.AgeSeconds = ageSeconds(txid, entry.Time, now)
		response.Message = "Transação aceita na mempool, aguardando inclusão em bloco."
		if response.AgeSeconds > 120 {
			response.Warning = "Transação está na mempool há mais de 2 minutos."
		}
		return response
	}

	if sentAt := trackedSentAt(txid); sentAt > 0 {
		response.Status = "broadcast"
		response.AgeSeconds = now - sentAt
		response.Message = "Transação enviada ao node, aguardando aceitação na mempool."
		return response
	}

	response.Warning = "Transação não localizada na wallet selecionada."
	return response
}

func trackedSentAt(txid string) int64 {
	trackedTxsMu.RLock()
	defer trackedTxsMu.RUnlock()

	return trackedTxs[txid].SentAt
}

func ageSeconds(txid string, fallbackUnix int64, now int64) int64 {
	if sentAt := trackedSentAt(txid); sentAt > 0 {
		return now - sentAt
	}

	if fallbackUnix > 0 {
		return now - fallbackUnix
	}

	return 0
}

func runNodeRPC(args ...string) ([]byte, error) {
	commandArgs := append(bitcoinCLIBaseArgs(), args...)
	cmd := exec.Command("bitcoin-cli", commandArgs...)
	return cmd.CombinedOutput()
}

func runWalletRPC(wallet string, args ...string) ([]byte, error) {
	commandArgs := append(bitcoinCLIBaseArgs(), "-rpcwallet="+wallet)
	commandArgs = append(commandArgs, args...)
	cmd := exec.Command("bitcoin-cli", commandArgs...)
	return cmd.CombinedOutput()
}

func bitcoinCLIBaseArgs() []string {
	args := []string{"-regtest"}

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

func btcToSats(amount float64) int64 {
	return int64(math.Round(amount * 100_000_000))
}

func satsToBTC(sats int64) float64 {
	return float64(sats) / 100_000_000
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}

	return false
}
