package main

type MempoolTx struct {
	VSize int `json:"vsize"`

	Fees struct {
		Base float64 `json:"base"`
	} `json:"fees"`
}

type MempoolSummary struct {
	TxCount         int            `json:"tx_count"`
	TotalVSize      int            `json:"total_vsize"`
	AvgFeeRate      float64        `json:"avg_fee_rate"`
	MinFeeRate      float64        `json:"min_fee_rate"`
	MaxFeeRate      float64        `json:"max_fee_rate"`
	FeeDistribution map[string]int `json:"fee_distribution"`
}

type BlockchainInfo struct {
	Blocks          int `json:"blocks"`
	Headers         int `json:"headers"`
	BlocksToHeaders int `json:"blocks_to_headers"`
}
