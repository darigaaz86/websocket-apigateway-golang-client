package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"
)

type Message struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

type Payload struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

const wsURL = "wss://l20rbivjj2.execute-api.ap-southeast-1.amazonaws.com/production/" // Replace this with your endpoint

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

	// Start pinging in background to keep connection alive
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

		fmt.Printf("📨 Message from %s: %s\n", msg.Payload.User, msg.Payload.Message)
	}
}
