package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"notiair/internal/routing"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
		token:   token,
		http:    &http.Client{},
	}
}

type sendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func (c *Client) SendMessage(ctx context.Context, chatID string, task routing.Task) error {
	text := fmt.Sprintf("Workflow %s\nTemplate %s\nPayload: %v", task.WorkflowID, task.TemplateID, task.Payload)

	body := sendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/sendMessage", strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var tgResp sendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return err
	}

	if !tgResp.OK {
		return fmt.Errorf("telegram api error: %s", tgResp.Description)
	}

	return nil
}
