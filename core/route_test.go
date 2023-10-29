package core_test

import (
	"context"
	"encoding/json"
	"errors"
	"magicrouter/core"
	"magicrouter/mocks"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFallbackChatService_ChatCompletion(t *testing.T) {
	// Happy path - Single route
	t.Run("happy path - single route", func(t *testing.T) {
		mockService := mocks.NewChatService(t)
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-3.5-turbo", "test").
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}, nil)
		svc := core.NewFallbackChatService(
			[]core.Route{
				{
					ID:            "route1",
					Priority:      1,
					Provider:      "openai",
					Model:         "gpt-3.5-turbo",
					ProviderToken: "test",
				},
			},
			core.ChatServices{
				"openai": mockService,
			},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	// Happy path - Multiple routes
	t.Run("happy path - multiple routes", func(t *testing.T) {
		mockService := mocks.NewChatService(t)
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-3.5-turbo", "test").
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}, nil).
			Once()
		svc := core.NewFallbackChatService(
			[]core.Route{
				{
					ID:            "route1",
					Priority:      1,
					Provider:      "openai",
					Model:         "gpt-3.5-turbo",
					ProviderToken: "test",
				},
				{
					ID:            "route2",
					Priority:      2,
					Provider:      "openai",
					Model:         "gpt-4",
					ProviderToken: "test",
				},
			},
			core.ChatServices{
				"openai": mockService,
			},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	// Sad path - First route fails with rate limit
	t.Run("sad path - first route fails with rate limit", func(t *testing.T) {
		mockService := mocks.NewChatService(t)
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-3.5-turbo", "test").
			Return(nil, core.ErrProviderRateLimited).
			Once()
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-4", "test").
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}, nil).
			Once()
		svc := core.NewFallbackChatService(
			[]core.Route{
				{
					ID:            "route1",
					Priority:      1,
					Provider:      "openai",
					Model:         "gpt-3.5-turbo",
					ProviderToken: "test",
				},
				{
					ID:            "route2",
					Priority:      2,
					Provider:      "openai",
					Model:         "gpt-4",
					ProviderToken: "test",
				},
			},
			core.ChatServices{
				"openai": mockService,
			},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	// Sad path - All routes fail
	t.Run("sad path - all routes fail", func(t *testing.T) {
		mockService := mocks.NewChatService(t)
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-3.5-turbo", "test").
			Return(nil, core.ErrProviderRateLimited).
			Once()
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-4", "test").
			Return(nil, core.ErrProviderTimeout).
			Once()
		svc := core.NewFallbackChatService(
			[]core.Route{
				{
					ID:            "route1",
					Priority:      1,
					Provider:      "openai",
					Model:         "gpt-3.5-turbo",
					ProviderToken: "test",
				},
				{
					ID:            "route2",
					Priority:      2,
					Provider:      "openai",
					Model:         "gpt-4",
					ProviderToken: "test",
				},
			},
			core.ChatServices{
				"openai": mockService,
			},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	// Sad path - First route fails with non-retryable error
	t.Run("sad path - first route fails with non-retryable error", func(t *testing.T) {
		mockService := mocks.NewChatService(t)
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-3.5-turbo", "test").
			Return(nil, errors.New("kaboom")).
			Once()
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-4", "test").
			Return(&http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
			}, nil).
			Once()
		svc := core.NewFallbackChatService(
			[]core.Route{
				{
					ID:            "route1",
					Priority:      1,
					Provider:      "openai",
					Model:         "gpt-3.5-turbo",
					ProviderToken: "test",
				},
				{
					ID:            "route2",
					Priority:      2,
					Provider:      "openai",
					Model:         "gpt-4",
					ProviderToken: "test",
				},
			},
			core.ChatServices{
				"openai": mockService,
			},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	// Sad path - All routes fail with non-retryable error
	t.Run("sad path - all routes fail with non-retryable error", func(t *testing.T) {
		mockService := mocks.NewChatService(t)
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-3.5-turbo", "test").
			Return(nil, errors.New("kaboom")).
			Once()
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-4", "test").
			Return(nil, errors.New("kaboom")).
			Once()
		svc := core.NewFallbackChatService(
			[]core.Route{
				{
					ID:            "route1",
					Priority:      1,
					Provider:      "openai",
					Model:         "gpt-3.5-turbo",
					ProviderToken: "test",
				},
				{
					ID:            "route2",
					Priority:      2,
					Provider:      "openai",
					Model:         "gpt-4",
					ProviderToken: "test",
				},
			},
			core.ChatServices{
				"openai": mockService,
			},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
