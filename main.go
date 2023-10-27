package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"magicrouter/magicrouter"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
)

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

type TokenStore interface {
	FetchProviderToken(apiToken string) (string, error)
}

type InMemTokenStore map[string]string

func (s InMemTokenStore) FetchProviderToken(apiToken string) (string, error) {
	token, ok := s[apiToken]
	if !ok {
		return "", errors.New("provider token not found")
	}
	return token, nil
}

type Server struct {
	client     *http.Client
	tokenStore TokenStore
}

func (s *Server) ChatCompletionHandler(w http.ResponseWriter, r *http.Request) {
	// We need to read the body twice, so let's keep it in a buffer.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Validate request
	var req magicrouter.ChatCompletionRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Fetch actual openai api key from DB
	apiToken := strings.Split(r.Header.Get("Authorization"), " ")[1]
	providerToken, err := s.tokenStore.FetchProviderToken(apiToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Proxy request
	hReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hReq.Header.Set("Authorization", fmt.Sprintf(" Bearer %s", providerToken))
	hReq.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(hReq)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

func (s *Server) ListenAndServe() error {
	r := chi.NewRouter()
	r.Post("/v1/chat/completions", s.ChatCompletionHandler)

	return http.ListenAndServe(":9200", r)
}

func main() {
	server := &Server{
		client:     &http.Client{},
		tokenStore: InMemTokenStore{"test": os.Getenv("OPENAI_API_KEY")},
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
