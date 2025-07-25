package api

import (
	"go-chat/internal/domain"
	"go-chat/internal/repository"
	"go-chat/internal/websocket"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type RoomHandler struct {
	roomRepo repository.RoomRepository
	hubManager *websocket.HubManager
}

func NewRoomHandler(roomRepo repository.RoomRepository, hubManager *websocket.HubManager) *RoomHandler {
	return &RoomHandler{roomRepo: roomRepo, hubManager: hubManager}
}

type CreateRoomRequest struct {
	Name string `json:"name" validate:"required,min=3,max=50"`
}

// CreateRoom обрабатывает создание новой комнаты чата.
func (h *RoomHandler) CreateRoom(c echo.Context) error {
	req := new(CreateRoomRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	room := &domain.Room{
		Name: req.Name,
	}

	if err := h.roomRepo.CreateRoom(c.Request().Context(), room); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create room"})
	}

	return c.JSON(http.StatusCreated, room)
}

// GetRooms обрабатывает получение всех доступных комнат чата.
func (h *RoomHandler) GetRooms(c echo.Context) error {
	rooms, err := h.roomRepo.GetRooms(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch rooms"})
	}

	return c.JSON(http.StatusOK, rooms)
}

type PostMessageRequest struct {
	Content string `json:"content" validate:"required,max=1000"`
}

// PostMessage обрабатывает отправку нового сообщения в определенную комнату.
func (h *RoomHandler) PostMessage(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid room ID"})
	}

	req := new(PostMessageRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*domain.JWTCustomClaims)
	userID := claims.UserID
	username := claims.Username

	message := &domain.Message{
		RoomID:  roomID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := h.roomRepo.SaveMessage(c.Request().Context(), message); err != nil {
		// В будущем здесь можно будет проверить ошибку внешнего ключа, чтобы убедиться, что комната существует.
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save message"})
	}

	// Добавляем имя пользователя в сообщение перед отправкой через WebSocket
	message.Username = username

	// Находим хаб для этой комнаты и отправляем сообщение, если хаб существует (т.е. есть подписчики)
	if hub, ok := h.hubManager.GetHub(roomID); ok {
		hub.Broadcast(message)
	}

	return c.JSON(http.StatusCreated, message)
}

// GetMessages обрабатывает получение всех сообщений для определенной комнаты.
func (h *RoomHandler) GetMessages(c echo.Context) error {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid room ID"})
	}

	messages, err := h.roomRepo.GetMessagesByRoomID(c.Request().Context(), roomID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch messages"})
	}

	return c.JSON(http.StatusOK, messages)
}