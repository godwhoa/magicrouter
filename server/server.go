package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sashabaranov/go-openai"
)

func getBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid authorization header")
	}
	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

func proxySSE(w http.ResponseWriter, body io.Reader) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	flusher := w.(http.Flusher)
	for scanner := bufio.NewScanner(body); scanner.Scan(); {
		w.Write(scanner.Bytes())
		w.Write([]byte("\n"))
		flusher.Flush()
	}
}

// TokenResolver resolves a provider and a provider token from an api token.
type TokenResolver interface {
	ResolveProviderToken(apiToken string) (provider string, pToken string, err error)
}

type ChatService interface {
	ChatCompletion(ctx context.Context, req json.RawMessage, token string) (*http.Response, error)
}

type Server struct {
	tokenStore TokenResolver
	services   map[string]ChatService
}

func New(tokenStore TokenResolver, services map[string]ChatService) *Server {
	return &Server{
		tokenStore: tokenStore,
		services:   services,
	}
}

func (s *Server) ChatCompletionHandler(w http.ResponseWriter, r *http.Request) {
	// We need to read the body twice, so let's keep it in a buffer.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Validate request
	var req openai.ChatCompletionRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Fetch actual openai api key from DB
	apiToken, err := getBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	provider, providerToken, err := s.tokenStore.ResolveProviderToken(apiToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	service, ok := s.services[provider]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := service.ChatCompletion(r.Context(), json.RawMessage(body), providerToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer io.Copy(io.Discard, response.Body)
	defer response.Body.Close()

	// Proxy response
	if response.Header.Get("Content-Type") == "text/event-stream" {
		proxySSE(w, response.Body)
		return
	}
	w.WriteHeader(response.StatusCode)
	_, err = io.Copy(w, response.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleError(fn func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *Server) ListenAndServe() error {
	r := chi.NewRouter()
	r.Post("/v1/chat/completions", s.ChatCompletionHandler)

	return http.ListenAndServe(":9200", r)
}
