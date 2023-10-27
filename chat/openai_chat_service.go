package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenAIChatService struct {
	client *http.Client
}

func NewOpenAIChatService(client *http.Client) *OpenAIChatService {
	return &OpenAIChatService{
		client: client,
	}
}

func (s *OpenAIChatService) ChatCompletion(ctx context.Context, req json.RawMessage, token string) (*http.Response, error) {
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
