package main

type WalletsResponse struct {
	AvailableWallets []string `json:"available_wallets"`
	LoadedWallets    []string `json:"loaded_wallets"`
	SelectedWallet   string   `json:"selected_wallet"`
}

type SelectWalletRequest struct {
	Wallet string `json:"wallet"`
}

type SelectWalletResponse struct {
	SelectedWallet string     `json:"selected_wallet"`
	WalletInfo     WalletInfo `json:"wallet_info"`
}

type WalletInfo struct {
	WalletName string  `json:"walletname"`
	Balance    float64 `json:"balance"`
	TxCount    int     `json:"txcount"`
}

type WalletStatusResponse struct {
	Wallet  string  `json:"wallet"`
	Balance float64 `json:"balance"`
	UTXOs   int     `json:"utxos"`
}

type ListWalletDirResponse struct {
	Wallets []WalletDirItem `json:"wallets"`
}

type WalletDirItem struct {
	Name string `json:"name"`
}

type SendTransactionRequest struct {
	Address string  `json:"address"`
	Amount  float64 `json:"amount"`
}

type SendTransactionResponse struct {
	TxID   string `json:"txid"`
	Wallet string `json:"wallet"`
	RawTx  string `json:"raw_tx"`
}

type UTXO struct {
	TxID          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Amount        float64 `json:"amount"`
	Confirmations int     `json:"confirmations"`
}

type RawTxInput struct {
	TxID string `json:"txid"`
	Vout int    `json:"vout"`
}

type SignRawTransactionResponse struct {
	Hex      string `json:"hex"`
	Complete bool   `json:"complete"`
}

type WalletTransaction struct {
	TxID          string `json:"txid"`
	Confirmations int    `json:"confirmations"`
	BlockHash     string `json:"blockhash"`
	Time          int64  `json:"time"`
}

type MempoolEntry struct {
	Time int64 `json:"time"`
}

type TrackedTransaction struct {
	TxID   string `json:"txid"`
	Wallet string `json:"wallet"`
	SentAt int64  `json:"sent_at"`
}

type TransactionResponse struct {
	TxID          string `json:"txid"`
	Wallet        string `json:"wallet"`
	Status        string `json:"status"`
	Confirmed     bool   `json:"confirmed"`
	Confirmations int    `json:"confirmations"`
	BlockHash     string `json:"block_hash"`
	AgeSeconds    int64  `json:"age_seconds"`
	Message       string `json:"message"`
	Warning       string `json:"warning,omitempty"`
}

type TransactionsResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
}
