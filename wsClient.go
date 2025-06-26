package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type ConnectionInfoPayload struct {
	OperationType string `json:"operationType"`
	ConnectionID  string `json:"connectionId"`
}

type Message struct {
	Action        string            `json:"action"`
	SourceId      string            `json:"sourceId"`
	CliToMpc      map[string]string `json:"cliToMpc"`
	OperationType string            `json:"operationType"` // e.g., "pairing", "signing"
	Message       json.RawMessage   `json:"message"`       // decode based on operationType
}

// Define payloads
type PairingPayload struct {
	DeviceID string `json:"deviceId"`
	User     string `json:"user"`
}

type SigningPayload struct {
	AccountHash   string `json:"accountHash"`
	TeamId        string `json:"teamId"`
	TransactionId string `json:"transactionId"`
	PartialSig    string `json:"partialSig"`
}

// WebSocket endpoint
const (
	rawURL             = "wss://eqm3whvj69.execute-api.ap-southeast-1.amazonaws.com/production?type=cli&cliId=cli123"
	allowInsecureTLS   = false // ⚠️ Set true only for development with self-signed certs
	connectionKeepTime = 4 * time.Minute
)

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid websocket URL: %w", err)
	}

	// Add query parameters
	// q := u.Query()
	// q.Set("type", "cli")
	// q.Set("cliId", "cli123")
	// u.RawQuery = q.Encode()

	// Setup custom TLS config (if needed)
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: allowInsecureTLS},
	}

	// Connect
	conn, _, err := dialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("WebSocket connect error: %w", err)
	}
	defer func() {
		_ = conn.Close()
		log.Println("❎ Connection closed")
	}()

	log.Println("✅ Connected to WebSocket server")

	// Keepalive pinger
	go func() {
		ticker := time.NewTicker(connectionKeepTime)
		defer ticker.Stop()

		for range ticker.C {
			err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(10*time.Second))
			if err != nil {
				log.Printf("⚠️ Ping failed: %v", err)
				return
			}
			log.Println("🔁 Ping sent to keep connection alive")
		}
	}()

	// Read loop
	for {
		_, data, err := conn.ReadMessage()
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

		case "PartialSig":
			var payload SigningPayload
			if err := json.Unmarshal(msg.Message, &payload); err != nil {
				log.Printf("❌ SigningPayload unmarshal error: %v", err)
				continue
			}
			fmt.Printf("✍️ Signing input: %v\n", payload)

			msg := map[string]string{
				"transactionId": payload.TransactionId,
				"teamId":        payload.TeamId,
				"accountHash":   payload.AccountHash,
				"signatureR":    "r",
				"signatureS":    "s",
				"signatureV":    "v",
			}
			msgBytes, _ := json.Marshal(msg)
			response := Message{
				Action:        "sendServer",
				SourceId:      "cli123",
				OperationType: "FullSig",
				Message: msgBytes,
			}

			if err := conn.WriteJSON(response); err != nil {
				log.Printf("❌ Failed to send FullSig: %v", err)
			}

		default:
			log.Printf("⚠️ Unknown operation type: %s", msg.OperationType)
		}
	}
}
