package openai

import (
	"context"
	"magicrouter/core"
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func runTestServer(t *testing.T) {
	r := chi.NewRouter()
	r.Post("/too-many-requests", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	r.Post("/ok", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Post("/timeout", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
	})
	go func() {
		err := http.ListenAndServe(":9252", r)
		assert.NoError(t, err)
	}()
}

func TestChatService_ChatCompletion(t *testing.T) {
	t.Parallel()
	runTestServer(t)

	tests := []struct {
		name     string
		endpoint string
		err      error
	}{
		{
			name:     "no error",
			endpoint: "http://localhost:9252/ok",
			err:      nil,
		},
		{
			name:     "timeout",
			endpoint: "http://localhost:9252/timeout",
			err:      core.ErrProviderTimeout,
		},
		{
			name:     "rate limited",
			endpoint: "http://localhost:9252/too-many-requests",
			err:      core.ErrProviderRateLimited,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &ChatService{
				client: &http.Client{
					Timeout: 500 * time.Millisecond,
				},
				endpoint: tt.endpoint,
			}
			_, err := svc.ChatCompletion(context.Background(), []byte(`{}`), "model", "token")
			assert.Equal(t, tt.err, err)
		})
	}

}
