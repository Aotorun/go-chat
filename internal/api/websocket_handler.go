package api

import (
	"go-chat/internal/domain"
	ws "go-chat/internal/websocket"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Для разработки разрешаем все источники.
			// В продакшене здесь должна быть проверка на ваш домен фронтенда.
			return true
		},
	}
)

type WebSocketHandler struct {
	hubManager *ws.HubManager
}

func NewWebSocketHandler(hubManager *ws.HubManager) *WebSocketHandler {
	return &WebSocketHandler{hubManager: hubManager}
}

// ServeWs обрабатывает WebSocket запросы.
func (h *WebSocketHandler) ServeWs(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid room ID"})
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v", err)
		return err
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*domain.JWTCustomClaims)
	userID := claims.UserID

	// Получаем или создаем хаб для конкретной комнаты
	hub := h.hubManager.GetOrCreateHub(roomID)

	client := &ws.Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan *domain.Message, 256),
		UserID: userID,
		RoomID: roomID,
	}
	client.Hub.Register(client)

	// Запускаем обработчики чтения и записи в отдельных горутинах.
	go client.WritePump()
	go client.ReadPump()

	return nil
}
