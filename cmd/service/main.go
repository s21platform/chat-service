package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/s21platform/chat-service/internal/config"
	db "github.com/s21platform/chat-service/internal/repository/postgres"
	"github.com/s21platform/chat-service/internal/service"
	"google.golang.org/grpc"

	chat "github.com/s21platform/chat-proto/chat-proto"
)

func main() {
	cfg := config.MustLoad()

	dbRepo, err := db.New(cfg)
	if err != nil {
		log.Printf("failed connection to db: %v", err)
		os.Exit(1)
	}

	defer dbRepo.Close()

	server := service.New(dbRepo)
	s := grpc.NewServer()

	chat.RegisterChatServiceServer(s, server)

	log.Println("starting chat service on port", cfg.Service.Port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Service.Port))
	if err != nil {
		log.Fatalf("Cannot listen port: %s; Error: %v", cfg.Service.Port, err)
	}
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Cannot start grpc: %s; Error: %v", cfg.Service.Port, err)
	}
}
