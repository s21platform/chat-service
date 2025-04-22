package main

import (
	"context"
	kafkalib "github.com/s21platform/kafka-lib"
	"github.com/s21platform/metrics-lib/pkg"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/databus/new_nickname"
	"github.com/s21platform/chat-service/internal/repository/postgres"
)

func main() {
	cfg := config.MustLoad()

	if cfg.Kafka.Host == "" {
		cfg.Kafka.Host = os.Getenv("KAFKA_HOST")
	}

	if cfg.Kafka.Port == "" {
		cfg.Kafka.Port = os.Getenv("KAFKA_PORT")
	}

	if cfg.Kafka.UpdateNickname == "" {
		cfg.Kafka.UpdateNickname = os.Getenv("KAFKA_UPDATE_NICKNAME")
	}

	if cfg.Metrics.Host == "" {
		cfg.Metrics.Host = os.Getenv("GRAFANA_HOST")
	}

	if cfg.Metrics.Port == 0 {
		portStr := os.Getenv("GRAFANA_PORT")
		if portStr != "" {
			port, err := strconv.Atoi(portStr)
			if err == nil {
				cfg.Metrics.Port = port
			}
		}
	}

	log.Printf("Postgres settings: host=%s, port=%s, user=%s, db=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Database)
	log.Printf("Kafka settings: host=%s, port=%s, topic=%s",
		cfg.Kafka.Host, cfg.Kafka.Port, cfg.Kafka.UpdateNickname)
	log.Printf("Metrics settings: host=%s, port=%d",
		cfg.Metrics.Host, cfg.Metrics.Port)

	if cfg.Kafka.Host == "" || cfg.Kafka.Port == "" {
		log.Fatal("Kafka host or port is not set")
	}

	if cfg.Kafka.UpdateNickname == "" {
		log.Fatal("Kafka topic for nickname updates is not set")
	}

	dbRepo := postgres.New(cfg)
	defer dbRepo.Close()

	metrics, err := pkg.NewMetrics(cfg.Metrics.Host, cfg.Metrics.Port, "chat", cfg.Platform.Env)
	if err != nil {
		log.Println("failed to connect graphite: ", err)
	}

	ctx := context.WithValue(context.Background(), config.KeyMetrics, metrics)

	consumerConfig := kafkalib.DefaultConsumerConfig(
		cfg.Kafka.Host,
		cfg.Kafka.Port,
		cfg.Kafka.UpdateNickname,
		"chat-nickname-updater",
	)

	nicknameConsumer, err := kafkalib.NewConsumer(consumerConfig, metrics)
	if err != nil {
		log.Fatalf("error create consumer: %v", err)
	}

	nicknameHandler := new_nickname.New(dbRepo)
	nicknameConsumer.RegisterHandler(ctx, func(ctx context.Context, msg []byte) error {
		nicknameHandler.Handler(ctx, msg)
		return nil
	})

	log.Println("Nickname consumer started successfully!")

	<-ctx.Done()
}
