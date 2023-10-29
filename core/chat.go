package core

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrProviderRateLimited = errors.New("provider rate limited")
	ErrProviderTimeout     = errors.New("provider timeout")
)

type ChatService interface {
	ChatCompletion(ctx context.Context, req json.RawMessage, model, token string) (*http.Response, error)
}

type ChatServices map[string]ChatService
