package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	zmq "github.com/go-zeromq/zmq4"
)

const zmqReconnectDelay = 5 * time.Second

func listen(ctx context.Context, address string, topic string, store *EventStore) {
	for {
		if err := listenOnce(ctx, address, topic, store); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			fmt.Println("zmq error:", err)
		}

		if !waitForReconnect(ctx) {
			return
		}
	}
}

func listenOnce(ctx context.Context, address string, topic string, store *EventStore) error {
	// O socket SUB recebe mensagens publicadas pelo Bitcoin Core via ZMQ.
	sub := zmq.NewSub(ctx)
	defer sub.Close()

	if err := sub.Dial(address); err != nil {
		return err
	}

	// SetSubscribe filtra para receber apenas o tópico desejado.
	if err := sub.SetOption(zmq.OptionSubscribe, topic); err != nil {
		return err
	}

	fmt.Println("listening", topic, "on", address)

	for {
		msg, err := sub.Recv()
		if err != nil {
			return err
		}

		if len(msg.Frames) < 2 {
			continue
		}

		eventTopic := string(msg.Frames[0])
		hashBytes := msg.Frames[1]
		observedAt := time.Now()

		store.Add(eventTopic, observedAt.Unix())
		hash := bitcoinHashToString(hashBytes)

		fmt.Printf("event=%s hash=%s observed_at=%s\n", eventTopic, hash, observedAt.Format(time.RFC3339))
	}
}

func bitcoinHashToString(b []byte) string {
	// O Bitcoin envia hashes em little-endian; inverter deixa igual ao formato exibido em exploradores.
	reversed := make([]byte, len(b))

	for i := range b {
		reversed[i] = b[len(b)-1-i]
	}

	return hex.EncodeToString(reversed)
}

func waitForReconnect(ctx context.Context) bool {
	timer := time.NewTimer(zmqReconnectDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
