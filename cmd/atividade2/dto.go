package main

import "sync"

type ObservedEvent struct {
	Topic      string
	Hash       string
	ObservedAt int64
}

type EventsSummary struct {
	BlocksObserved int     `json:"blocks_observed"`
	TxObserved     int     `json:"tx_observed"`
	LastEventTime  int64   `json:"last_event_time"`
	TxPerSecond    float64 `json:"tx_per_second"`
}

type LatestBlockEvent struct {
	Hash string `json:"hash"`
	Ts   int64  `json:"ts"`
}

type LatestTxEvent struct {
	TxID string `json:"txid"`
	Ts   int64  `json:"ts"`
}

type LatestEvents struct {
	Blocks []LatestBlockEvent `json:"blocks"`
	Txs    []LatestTxEvent    `json:"txs"`
}

type EventStore struct {
	mu     sync.RWMutex
	limit  int
	events []ObservedEvent
}
