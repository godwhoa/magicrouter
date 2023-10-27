package providers

import (
	"context"
	"encoding/json"
	"net/http"
)

type ChatService interface {
	ChatCompletion(ctx context.Context, req json.RawMessage, token string) (*http.Response, error)
}
