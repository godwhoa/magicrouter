package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"magicrouter/core"

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
	tokenResolver core.TokenResolver
	services      core.ChatServices
	projectStore  core.ProjectStore
}

func New(tokenStore core.TokenResolver, services core.ChatServices, projectStore core.ProjectStore) *Server {
	return &Server{
		tokenResolver: tokenStore,
		services:      services,
		projectStore:  projectStore,
	}
}

func (s *Server) ChatCompletionHandler(w http.ResponseWriter, r *http.Request) error {
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

	// Get project config which contains fallback configuration
	projectID := getProjectID(r.Context())
	cfg, err := s.projectStore.GetConfig(projectID)
	if err != nil {
		return fmt.Errorf("failed to get project config: %w", err)
	}

	// Send request to provider
	service := core.NewFallbackChatService(cfg.Routes, s.services)
	response, err := service.ChatCompletion(r.Context(), json.RawMessage(body))
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
