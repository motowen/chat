package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer c.Close()

	// 設定 pong handler
	c.SetPongHandler(func(appData string) error {
		log.Println("received pong from server:", appData)
		return nil
	})

	// 定期發送 ping
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := c.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				log.Println("Ping error:", err)
				return
			}
		}
	}()

	// 讀取 server 訊息
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("recv: %s", message)
	}
}
