package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
)

type Route struct {
	ID            string
	Priority      int
	Provider      string
	Model         string
	ProviderToken string
}

type FallbackChatService struct {
	routes   []Route
	services ChatServices
}

func NewFallbackChatService(routes []Route, services ChatServices) *FallbackChatService {
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Priority > routes[j].Priority
	})
	return &FallbackChatService{
		routes:   routes,
		services: services,
	}
}

func (s *FallbackChatService) ChatCompletion(ctx context.Context, req json.RawMessage) (*http.Response, error) {
	for _, route := range s.routes {
		svc, ok := s.services[route.Provider]
		if !ok {
			return nil, fmt.Errorf("unknown provider: %s", route.Provider)
		}

		resp, err := svc.ChatCompletion(ctx, req, route.Model, route.ProviderToken)
		if errors.Is(err, ErrProviderRateLimited) || errors.Is(err, ErrProviderTimeout) {
			continue
		}
		if err != nil {
			return nil, err
		}

		return resp, nil
	}

	return nil, fmt.Errorf("all providers rate limited or timed out")
}
