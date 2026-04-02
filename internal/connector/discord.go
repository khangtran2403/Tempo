package connector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brokeboycoding/tempo/internal/storage"
)

type DiscordConnector struct {
	integrationRepo storage.IntegrationRepository
	encryptionKey   string
}

func NewDiscordConnector(
	integrationRepo storage.IntegrationRepository,
	encryptionKey string,
) *DiscordConnector {
	return &DiscordConnector{
		integrationRepo: integrationRepo,
		encryptionKey:   encryptionKey,
	}
}

func (dc *DiscordConnector) Name() string {
	return "discord"
}

func (dc *DiscordConnector) Execute(
	ctx context.Context,
	config map[string]interface{},
	prevResults map[string]interface{},
) (interface{}, error) {
	action, _ := config["action"].(string)

	switch action {
	case "send_message":
		return dc.sendMessage(ctx, config)
	case "send_embed":
		return dc.sendEmbed(ctx, config)
	default:
		return nil, fmt.Errorf("hành động không tồn tại: %s", action)
	}
}

func (dc *DiscordConnector) sendMessage(
	ctx context.Context,
	config map[string]interface{},
) (interface{}, error) {
	webhookURL, ok := config["webhook_url"].(string)
	if !ok {
		return nil, fmt.Errorf("Không có URL webhook")
	}

	content, _ := config["content"].(string)
	username, _ := config["username"].(string)
	avatarURL, _ := config["avatar_url"].(string)

	payload := map[string]interface{}{
		"content":    content,
		"username":   username,
		"avatar_url": avatarURL,
	}

	data, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		webhookURL,
		bytes.NewReader(data),
	)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("lỗi discord api : trạng thái %d", resp.StatusCode)
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

func (dc *DiscordConnector) sendEmbed(
	ctx context.Context,
	config map[string]interface{},
) (interface{}, error) {
	webhookURL, _ := config["webhook_url"].(string)
	title, _ := config["title"].(string)
	description, _ := config["description"].(string)
	color, _ := config["color"].(float64)

	embed := map[string]interface{}{
		"title":       title,
		"description": description,
		"color":       int(color),
	}

	payload := map[string]interface{}{
		"embeds": []interface{}{embed},
	}

	data, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(
		ctx,
		"POST",
		webhookURL,
		bytes.NewReader(data),
	)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return map[string]interface{}{
		"success": true,
	}, nil
}

func (dc *DiscordConnector) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["webhook_url"].(string); !ok {
		return fmt.Errorf("thiếu webhook_url trong cấu hình")
	}
	return nil
}
