package websocket

import (
	"log"
	"sync"
)

// HubManager управляет всеми хабами для разных комнат.
type HubManager struct {
	hubs map[int64]*Hub
	mu   sync.RWMutex
}

func NewHubManager() *HubManager {
	return &HubManager{
		hubs: make(map[int64]*Hub),
	}
}

// GetOrCreateHub получает хаб для данного roomID, создавая его, если он не существует.
func (m *HubManager) GetOrCreateHub(roomID int64) *Hub {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hub, ok := m.hubs[roomID]; ok {
		return hub
	}

	hub := NewHub(roomID, m)
	go hub.Run()
	m.hubs[roomID] = hub
	log.Printf("Создан новый хаб для комнаты %d", roomID)
	return hub
}

// GetHub получает хаб, если он существует, иначе возвращает nil.
func (m *HubManager) GetHub(roomID int64) (*Hub, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	hub, ok := m.hubs[roomID]
	return hub, ok
}

// DeleteHub удаляет хаб из менеджера.
func (m *HubManager) DeleteHub(roomID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.hubs[roomID]; ok {
		delete(m.hubs, roomID)
		log.Printf("Хаб для комнаты %d удален", roomID)
	}
}
