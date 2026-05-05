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

type ListWalletDirResponse struct {
	Wallets []WalletDirItem `json:"wallets"`
}

type WalletDirItem struct {
	Name string `json:"name"`
}
