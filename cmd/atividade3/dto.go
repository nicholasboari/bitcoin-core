package main

type WalletsResponse struct {
	AvailableWallets []string `json:"available_wallets"`
	LoadedWallets    []string `json:"loaded_wallets"`
	SelectedWallet   string   `json:"selected_wallet"`
}

type ListWalletDirResponse struct {
	Wallets []WalletDirItem `json:"wallets"`
}

type WalletDirItem struct {
	Name string `json:"name"`
}
