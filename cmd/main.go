package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go-chat/internal/api"
	"go-chat/internal/config"
	"go-chat/internal/repository"
	"go-chat/internal/validator"
	"go-chat/internal/websocket"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	var dbpool *pgxpool.Pool
	for i := 0; i < 5; i++ {
		fmt.Printf("Попытка подключения к БД #%d...\n", i+1)
		dbpool, err = pgxpool.New(context.Background(), cfg.DBSource)
		if err == nil {
			if err = dbpool.Ping(context.Background()); err == nil {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Не удалось подключиться к базе данных после нескольких попыток: %v\n", err)
		os.Exit(1)
	}

	defer dbpool.Close()
	fmt.Println("Успешно подключились к базе данных!")

	userRepo := repository.NewUserRepository(dbpool)
	roomRepo := repository.NewRoomRepository(dbpool)

	// Создаем менеджер хабов
	hubManager := websocket.NewHubManager()
	fmt.Println("WebSocket Hub Manager запущен.")

	userHandler := api.NewUserHandler(userRepo, cfg)
	roomHandler := api.NewRoomHandler(roomRepo, hubManager)
	wsHandler := api.NewWebSocketHandler(hubManager)

	e := echo.New()
	e.Validator = validator.NewValidator()

	apiV1 := e.Group("/api/v1")

	// Публичные маршруты
	apiV1.POST("/register", userHandler.Register)
	apiV1.POST("/login", userHandler.Login)

	// Защищенные маршруты
	protected := apiV1.Group("")
	protected.Use(api.JWTMiddleware(cfg))
	
	protected.GET("/me", userHandler.Me)

	// Маршруты для комнат (защищенные)
	protected.POST("/rooms", roomHandler.CreateRoom)
	protected.GET("/rooms", roomHandler.GetRooms)
	protected.POST("/rooms/:id/messages", roomHandler.PostMessage)
	protected.GET("/rooms/:id/messages", roomHandler.GetMessages)

	// Маршрут для WebSocket
	protected.GET("/ws/rooms/:id", wsHandler.ServeWs)

	// Раздача статических файлов из папки public
	e.Static("/", "public")

	e.Logger.Fatal(e.Start(cfg.ServerAddress))
}
