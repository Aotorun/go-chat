package domain

import "time"

type Room struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}