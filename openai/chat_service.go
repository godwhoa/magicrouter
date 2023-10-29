package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"magicrouter/core"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/tidwall/sjson"
)

type ChatService struct {
	client   *http.Client
	endpoint string
}

func NewChatService(client *http.Client) *ChatService {
	return &ChatService{
		endpoint: "https://api.openai.com/v1/chat/completions",
		client:   client,
	}
}

func (s *ChatService) ChatCompletion(ctx context.Context, req json.RawMessage, model, token string) (*http.Response, error) {
	log.Info().Msg("openai.ChatCompletion")
	req, err := sjson.SetBytes(req, "model", model)
	if err != nil {
		return nil, fmt.Errorf("failed to update model: %w", err)
	}

	hReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewBuffer(req))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	hReq.Header.Set("Authorization", fmt.Sprintf(" Bearer %s", token))
	hReq.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(hReq)
	if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
		return nil, core.ErrProviderTimeout
	}
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	if response.StatusCode == http.StatusTooManyRequests {
		return nil, core.ErrProviderRateLimited
	}

	return response, nil
}
