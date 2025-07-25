package websocket

import (
	"log"
	"time"

	"go-chat/internal/domain"

	"github.com/gorilla/websocket"
)

const (
	// Время на запись сообщения.
	writeWait = 10 * time.Second
	// Время на чтение следующего pong-сообщения.
	pongWait = 60 * time.Second
	// Период отправки ping-сообщений. Должен быть меньше pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Максимальный размер сообщения.
	maxMessageSize = 1024
)

// Client - это посредник между WebSocket-соединением и хабом.
type Client struct {
	Hub *Hub
	// WebSocket-соединение.
	Conn *websocket.Conn
	// Буферизированный канал исходящих сообщений.
	Send chan *domain.Message
	// ID пользователя из JWT.
	UserID int64
	// ID комнаты, к которой подключен клиент.
	RoomID int64
}

// readPump считывает сообщения из WebSocket и передает их в хаб.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// В текущей реализации мы не принимаем сообщения от клиента через WebSocket,
	// так как они отправляются через REST API. Этот цикл нужен для поддержания
	// соединения и обработки его закрытия.
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}
	}
}

// writePump передает сообщения из хаба в WebSocket.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Канал был закрыт хабом.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("error writing json: %v", err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}