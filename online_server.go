package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: 實際專案應限制允許的 domain
	},
}

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	// 設定 pong handler
	conn.SetPongHandler(func(appData string) error {
		log.Println("received pong from client:", appData)
		return nil
	})

	// 啟動一個 goroutine 定期發送 ping
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				log.Println("Ping error:", err)
				return
			}
		}
	}()

	// 持續讀取 client 消息
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		fmt.Printf("recv: %s\n", msg)

		// 回應 echo
		if err := conn.WriteMessage(msgType, msg); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

func main() {
	r := gin.Default()
	r.GET("/ws", func(c *gin.Context) {
		wsHandler(c)
	})
	r.Run(":8080")
}
