package repository

import (
	"context"
	"go-chat/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RoomRepository определяет интерфейс для работы с комнатами и сообщениями.
type RoomRepository interface {
	CreateRoom(ctx context.Context, room *domain.Room) error
	GetRooms(ctx context.Context) ([]domain.Room, error)
	SaveMessage(ctx context.Context, message *domain.Message) error
	GetMessagesByRoomID(ctx context.Context, roomID int64) ([]domain.Message, error)
}

type pgxRoomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) RoomRepository {
	return &pgxRoomRepository{db: db}
}

func (r *pgxRoomRepository) CreateRoom(ctx context.Context, room *domain.Room) error {
	query := `INSERT INTO rooms (name) VALUES ($1) RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query, room.Name).Scan(&room.ID, &room.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *pgxRoomRepository) GetRooms(ctx context.Context) ([]domain.Room, error) {
	query := `SELECT id, name, created_at FROM rooms ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}

func (r *pgxRoomRepository) SaveMessage(ctx context.Context, message *domain.Message) error {
	query := `INSERT INTO messages (room_id, user_id, content) 
	          VALUES ($1, $2, $3) 
			  RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query, message.RoomID, message.UserID, message.Content).Scan(&message.ID, &message.CreatedAt)
	return err
}

func (r *pgxRoomRepository) GetMessagesByRoomID(ctx context.Context, roomID int64) ([]domain.Message, error) {
	query := `SELECT m.id, m.room_id, m.user_id, u.username, m.content, m.created_at 
	          FROM messages m
			  JOIN users u ON m.user_id = u.id
			  WHERE m.room_id = $1 
			  ORDER BY m.created_at ASC`
	rows, err := r.db.Query(ctx, query, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Username, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}