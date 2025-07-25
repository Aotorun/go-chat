package websocket

import (
	"go-chat/internal/domain"
)

// Hub поддерживает набор активных клиентов и рассылает им сообщения.
type Hub struct {
	// Зарегистрированные клиенты.
	clients map[*Client]bool

	// Входящие сообщения от клиентов для рассылки.
	broadcast chan *domain.Message

	// Канал для регистрации клиентов.
	register chan *Client

	// Канал для отмены регистрации клиентов.
	unregister chan *Client

	// ID комнаты, которую обслуживает этот хаб.
	RoomID int64

	// Менеджер хабов, чтобы хаб мог сам себя удалить.
	manager *HubManager
}

func NewHub(roomID int64, manager *HubManager) *Hub {
	return &Hub{
		broadcast:  make(chan *domain.Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		RoomID:     roomID,
		manager:    manager,
	}
}

// Broadcast отправляет сообщение в канал broadcast.
func (h *Hub) Broadcast(message *domain.Message) {
	h.broadcast <- message
}

// Register регистрирует нового клиента в хабе.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister отменяет регистрацию клиента.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				// Если в комнате не осталось клиентов, удаляем хаб.
				if len(h.clients) == 0 {
					h.manager.DeleteHub(h.RoomID)
				}
			}
		case message := <-h.broadcast:
			// Рассылаем сообщение всем клиентам, подключенным к этому хабу (комнате).
			for client := range h.clients {
				// Неблокирующая отправка, чтобы один медленный клиент не тормозил всех остальных.
				select {
				case client.Send <- message:
				default:
					// Если буфер клиента переполнен, закрываем его соединение.
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}
