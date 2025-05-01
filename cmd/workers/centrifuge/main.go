package main

import (
	"github.com/s21platform/chat-service/internal/config"
	centrifugeWorker "github.com/s21platform/chat-service/internal/workers/centrifuge"
	logger_lib "github.com/s21platform/logger-lib"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	logger := logger_lib.New(cfg.Logger.Host, cfg.Logger.Port, "centrifuge-worker", cfg.Platform.Env)
	logger.Info("Starting Centrifuge worker")

	worker := centrifugeWorker.New(logger)

	port := os.Getenv("CENTRIFUGE_PORT")
	if port == "" {
		port = "8093"
	}

	err := worker.Start(port)
	if err != nil {
		logger.Error("Failed to start Centrifuge worker: " + err.Error())
		os.Exit(1)
	}

	logger.Info("Centrifuge worker started successfully")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	logger.Info("Shutting down Centrifuge worker")

	err = worker.Stop()
	if err != nil {
		logger.Error("Error stopping Centrifuge worker: " + err.Error())
		os.Exit(1)
	}

	logger.Info("Centrifuge worker stopped gracefully")
}
