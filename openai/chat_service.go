package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tidwall/sjson"
)

type ChatService struct {
	client *http.Client
}

func NewChatService(client *http.Client) *ChatService {
	return &ChatService{
		client: client,
	}
}

func (s *ChatService) ChatCompletion(ctx context.Context, req json.RawMessage, model, token string) (*http.Response, error) {
	req, err := sjson.SetBytes(req, "model", model)
	if err != nil {
		return nil, fmt.Errorf("failed to update model: %w", err)
	}

	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(req))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	hReq.Header.Set("Authorization", fmt.Sprintf(" Bearer %s", token))
	hReq.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(hReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return response, nil
}
