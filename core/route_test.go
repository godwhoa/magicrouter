package core_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"magicrouter/core"
	"magicrouter/mocks"

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
			core.NoOpBreaker{},
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
			core.NoOpBreaker{},
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
			core.NoOpBreaker{},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	// Sad path - All routes fail
	t.Run("sad path - all routes fail - ensure FallbackError", func(t *testing.T) {
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
			core.NoOpBreaker{},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.Equal(t, core.FallbackError{
			"route1": core.ErrProviderRateLimited,
			"route2": core.ErrProviderTimeout,
		}, err)
		assert.Equal(t, err.Error(), "all routes failed: route1: provider rate limited, route2: provider timeout, ")
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
			core.NoOpBreaker{},
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
			core.NoOpBreaker{},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
	// Sad path - First and only route fails with rate limit, it should return the error
	t.Run("sad path - last route fails with rate limit, it should return the error", func(t *testing.T) {
		mockService := mocks.NewChatService(t)
		mockService.
			On("ChatCompletion", mock.Anything, json.RawMessage(`{"prompt": "hello"}`), "gpt-3.5-turbo", "test").
			Return(nil, core.ErrProviderRateLimited).
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
			},
			core.ChatServices{
				"openai": mockService,
			},
			core.NoOpBreaker{},
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("sad path - breaker open", func(t *testing.T) {
		mockService := mocks.NewChatService(t)
		mockBreaker := mocks.NewBreakerService(t)
		mockBreaker.On("GetState", mock.Anything, "route1").Return(core.BreakerStateOpen, nil).Once()
		mockBreaker.On("GetState", mock.Anything, "route2").Return(core.BreakerStateClosed, nil).Once()
		mockBreaker.On("ReportSuccess", mock.Anything, "route2").Return(nil).Once()
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
			mockBreaker,
		)
		resp, err := svc.ChatCompletion(context.Background(), json.RawMessage(`{"prompt": "hello"}`))
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
