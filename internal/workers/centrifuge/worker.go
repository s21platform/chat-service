package centrifuge

import (
	"context"
	"fmt"
	"github.com/centrifugal/centrifuge"
	logger_lib "github.com/s21platform/logger-lib"
	"net/http"
	"time"
)

type Worker struct {
	node       *centrifuge.Node
	logger     *logger_lib.Logger
	httpServer *http.Server
}

func New(logger *logger_lib.Logger) *Worker {
	return &Worker{
		logger: logger,
	}
}

func (w *Worker) Start(port string) error {
	cfg := centrifuge.Config{
		LogLevel:   centrifuge.LogLevelDebug,
		LogHandler: w.handleLog,
	}

	node, err := centrifuge.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create centrifuge node: %w", err)
	}
	w.node = node

	w.setupHandlers()

	err = node.Run()
	if err != nil {
		return fmt.Errorf("failed to run centrifuge node: %w", err)
	}

	wsHandler := centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{
		CheckOrigin: func(r *http.Request) bool {
			return true //TODO: В продакшене нужно настроить проверку origin
		},
	})

	mux := http.NewServeMux()
	mux.Handle("/connection/websocket", wsHandler)

	addr := fmt.Sprintf(":%s", port)
	w.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		w.logger.Info(fmt.Sprintf("Starting Centrifuge WebSocket server on %s", addr))
		err := w.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			w.logger.Error(fmt.Sprintf("Centrifuge HTTP server error: %v", err))
		}
	}()

	return nil
}

func (w *Worker) Stop() error {
	if w.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := w.httpServer.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("failed to shutdown HTTP server: %w", err)
		}
	}

	if w.node != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := w.node.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("failed to shutdown centrifuge node: %w", err)
		}
	}

	return nil
}

// setupHandlers настраивает обработчики событий для Centrifuge
func (w *Worker) setupHandlers() {
	w.node.OnConnecting(func(ctx context.Context, e centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
		w.logger.Info(fmt.Sprintf("Client connecting with token: %s", e.Token))

		cred := &centrifuge.Credentials{
			UserID: e.Token,
		}

		return centrifuge.ConnectReply{
			Credentials: cred,
		}, nil
	})

	// Обработчик события когда клиент успешно подключился
	w.node.OnConnect(func(client *centrifuge.Client) {
		w.logger.Info(fmt.Sprintf("Client connected: %s", client.ID()))

		// Обработчик подписки на канал
		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			w.logger.Info(fmt.Sprintf("Client %s subscribes to channel %s", client.ID(), e.Channel))

			// Для базового каркаса разрешаем подписку без дополнительных проверок
			cb(centrifuge.SubscribeReply{}, nil)
		})

		// Добавляем обработчик публикации
		client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
			w.logger.Info(fmt.Sprintf("Client %s publishes to channel %s: %s",
				client.ID(), e.Channel, string(e.Data)))

			// Разрешаем публикацию
			cb(centrifuge.PublishReply{}, nil)
		})

		// Обработчик отписки от канала
		client.OnUnsubscribe(func(e centrifuge.UnsubscribeEvent) {
			w.logger.Info(fmt.Sprintf("Client %s unsubscribed from channel %s", client.ID(), e.Channel))
		})

		// Обработчик отключения клиента
		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			w.logger.Info(fmt.Sprintf("Client %s disconnected, reason: %s", client.ID(), e.Disconnect.Reason))
		})
	})
}

func (w *Worker) Publish(channel string, data []byte) error {
	if w.node == nil {
		return fmt.Errorf("node not initialized")
	}

	_, err := w.node.Publish(channel, data)
	if err != nil {
		return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
	}

	return nil
}

func (w *Worker) handleLog(entry centrifuge.LogEntry) {
	switch entry.Level {
	case centrifuge.LogLevelError:
		w.logger.Error(entry.Message)
	case centrifuge.LogLevelInfo:
		w.logger.Info(entry.Message)
	}
}
