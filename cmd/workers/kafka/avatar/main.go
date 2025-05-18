package main

import (
	"context"
	"fmt"

	_ "github.com/lib/pq"

	kafkalib "github.com/s21platform/kafka-lib"
	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/metrics-lib/pkg"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/databus/avatar"
	"github.com/s21platform/chat-service/internal/repository/postgres"
)

const newAvatarConsumerGroupID = "avatar-updater"

func main() {
	cfg := config.MustLoad()
	logger := logger_lib.New(cfg.Logger.Host, cfg.Logger.Port, cfg.Service.Name, cfg.Platform.Env)

	dbRepo := postgres.New(cfg)
	defer dbRepo.Close()

	metrics, err := pkg.NewMetrics(cfg.Metrics.Host, cfg.Metrics.Port, cfg.Service.Name, cfg.Platform.Env)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect graphite: %v", err))
	}

	ctx := context.WithValue(context.Background(), config.KeyMetrics, metrics)
	ctx = context.WithValue(ctx, config.KeyLogger, logger)

	consumerConfig := kafkalib.DefaultConsumerConfig(
		cfg.Kafka.Host,
		cfg.Kafka.Port,
		cfg.Kafka.AvatarTopic,
		newAvatarConsumerGroupID,
	)
	consumer, err := kafkalib.NewConsumer(consumerConfig, metrics)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create consumer: %v", err))
	}

	avatarHandler := avatar.New(dbRepo)
	consumer.RegisterHandler(ctx, avatarHandler.Handler)

	<-ctx.Done()
}
