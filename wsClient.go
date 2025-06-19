package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"
)

type ConnectionInfoPayload struct {
	OperationType string `json:"operationType"`
	ConnectionID  string `json:"connectionId"`
}

type Message struct {
	Action        string          `json:"action"`
	ConnectionID  string          `json:"connectionId"`
	OperationType string          `json:"operationType"` // e.g., "pairing", "signing"
	Message       json.RawMessage `json:"message"`       // decode based on operationType
}

// Define payloads

type PairingPayload struct {
	DeviceID string `json:"deviceId"`
	User     string `json:"user"`
}

type SigningPayload struct {
	TxID      string `json:"txId"`
	Signature string `json:"signature"`
}

const wsURL = "wss://izu0a6unlg.execute-api.ap-southeast-1.amazonaws.com/production/"

func main() {
	for {
		err := connectAndListen()
		if err != nil {
			log.Printf("🔄 Reconnecting in 5 seconds due to error: %v", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func connectAndListen() error {
	ctx := context.Background()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("WebSocket connect error: %w", err)
	}
	defer func() {
		_ = conn.Close(websocket.StatusNormalClosure, "closing connection")
		log.Println("❎ Connection closed")
	}()

	log.Println("✅ Connected to WebSocket server")

	// 🔁 Keep connection alive
	go func() {
		ticker := time.NewTicker(4 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			err := conn.Ping(ctx)
			if err != nil {
				log.Printf("⚠️ Ping failed: %v", err)
				return
			}
			log.Println("🔁 Ping sent to keep connection alive")
		}
	}()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("❌ JSON unmarshal error: %v", err)
			continue
		}

		switch msg.OperationType {
		case "pairing":
			var payload PairingPayload
			if err := json.Unmarshal(msg.Message, &payload); err != nil {
				log.Printf("❌ PairingPayload unmarshal error: %v", err)
				continue
			}
			fmt.Printf("🔗 Pairing: device=%s user=%s\n", payload.DeviceID, payload.User)

		case "signing":
			var payload SigningPayload
			if err := json.Unmarshal(msg.Message, &payload); err != nil {
				log.Printf("❌ SigningPayload unmarshal error: %v", err)
				continue
			}
			fmt.Printf("✍️ Signing: txId=%s signature=%s\n", payload.TxID, payload.Signature)

		default:
			log.Printf("⚠️ Unknown operation type: %s", msg.OperationType)
		}
	}
}
