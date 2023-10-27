package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"magicrouter/chat"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
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

type Server struct {
	tokenResolver TokenResolver
	services      chat.Services
}

func New(tokenStore TokenResolver, services chat.Services) *Server {
	return &Server{
		tokenResolver: tokenStore,
		services:      services,
	}
}

func (s *Server) ChatCompletionHandler(w http.ResponseWriter, r *http.Request) error {
	providerContext := getProviderContext(r.Context())
	provider, providerToken := providerContext.provider, providerContext.providerToken

	// We need to read the body twice, so let's keep it in a slice.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Validate request
	var req openai.ChatCompletionRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		return HTTPError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid request body",
			Err:        err,
		}
	}

	// Get response from provider
	service, ok := s.services[provider]
	if !ok {
		return fmt.Errorf("unknown provider: %s", provider)
	}

	response, err := service.ChatCompletion(r.Context(), json.RawMessage(body), providerToken)
	if err != nil {
		return fmt.Errorf("service request failed: %w", err)
	}
	defer io.Copy(io.Discard, response.Body)
	defer response.Body.Close()

	// Proxy provider response
	if response.Header.Get("Content-Type") == "text/event-stream" {
		proxySSE(w, response.Body)
		return nil
	}

	w.WriteHeader(response.StatusCode)
	_, err = io.Copy(w, response.Body)
	if err != nil {
		return fmt.Errorf("failed to copy response body: %w", err)
	}
	return nil
}

func handleError(fn func(http.ResponseWriter, *http.Request) error) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(w, r)
		if err == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Error().Err(err).Msg("request failed")
		httpErr, ok := err.(HTTPError)
		if ok {
			w.WriteHeader(httpErr.StatusCode)
			if httpErr.Message != "" {
				json.NewEncoder(w).Encode(httpErr)
			}
			return
		}
	}
}

func (s *Server) ListenAndServe() error {
	r := chi.NewRouter()
	r.Use(requestLogger(log.Logger))
	r.Group(func(r chi.Router) {
		r.Use(resolveToken(s.tokenResolver))
		r.Post("/v1/chat/completions", handleError(s.ChatCompletionHandler))
	})

	return http.ListenAndServe(":9200", r)
}
