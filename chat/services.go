package chat

import (
	"context"
	"encoding/json"
	"net/http"
)

type Service interface {
	ChatCompletion(ctx context.Context, req json.RawMessage, token string) (*http.Response, error)
}

type Services = map[string]Service

func NewServiceMap() Services {
	return Services{
		"openai": NewOpenAIChatService(http.DefaultClient),
	}
}
