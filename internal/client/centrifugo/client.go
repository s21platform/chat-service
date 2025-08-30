package centrifugo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/s21platform/chat-service/internal/config"
	"github.com/s21platform/chat-service/internal/model"
)

const (
	publishMethod = "publish"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.Centrifuge.BaseURL,
		apiKey:  cfg.Centrifuge.APIKey,
		httpClient: &http.Client{
			Timeout: cfg.Centrifuge.Timeout,
		},
	}
}

func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}

func (c *Client) Publish(ctx context.Context, channel string, msg model.Message) error {
	payload := model.CentrifugoEvent{
		Method: publishMethod,
		Params: model.CentrifugoEventParams{
			Channel: channel,
			Data:    msg,
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "apikey "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // .

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if errorData, exists := response["error"]; exists && errorData != nil {
		return fmt.Errorf("centrifugo error: %v", errorData)
	}

	return nil
}
