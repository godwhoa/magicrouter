package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

type FallbackError map[string]error

func (e FallbackError) Error() string {
	var msg strings.Builder
	msg.WriteString("all routes failed: ")
	for route, err := range e {
		msg.WriteString(fmt.Sprintf("%s: %s, ", route, err.Error()))
	}
	return msg.String()
}

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
	breaker  BreakerService
}

func NewFallbackChatService(routes []Route, services ChatServices, breaker BreakerService) *FallbackChatService {
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Priority < routes[j].Priority // ascending
	})
	return &FallbackChatService{
		routes:   routes,
		services: services,
		breaker:  breaker,
	}
}

func (s *FallbackChatService) ChatCompletion(ctx context.Context, req json.RawMessage) (*http.Response, error) {
	fallbackErr := make(FallbackError)
	for _, route := range s.routes {
		state, err := s.breaker.GetState(ctx, route.ID)
		if err != nil {
			log.Err(err).Msg("failed to get breaker state")
		}
		if err == nil && !state.ShouldAttempt() {
			continue
		}

		svc, ok := s.services[route.Provider]
		if !ok {
			return nil, fmt.Errorf("unknown provider: %s", route.Provider)
		}

		resp, err := svc.ChatCompletion(ctx, req, route.Model, route.ProviderToken)
		if err != nil {
			fallbackErr[route.ID] = err
			s.breaker.ReportFailure(ctx, route.ID)
			continue
		}
		s.breaker.ReportSuccess(ctx, route.ID)
		return resp, nil
	}

	return nil, fallbackErr
}
