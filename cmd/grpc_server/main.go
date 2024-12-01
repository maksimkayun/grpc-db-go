package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"grpc/internal/config"
	"grpc/internal/service" // добавьте импорт вашего сервиса
	desc "grpc/pkg/note_v1"
)

const grpcPort = 5001

func main() {
	cfg := config.Load()

	ctx := context.Background()
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	// Инициализируем сервис с пулом подключений
	noteService := service.NewNoteService(pool)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)

	// Регистрируем наш сервис вместо заглушки
	desc.RegisterNoteV1Server(s, noteService)

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
