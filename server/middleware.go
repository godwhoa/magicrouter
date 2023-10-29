package server

import (
	"context"
	"magicrouter/core"
	"net/http"
	"runtime/debug"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type providerContext struct {
	provider      string
	providerToken string
	apiToken      string
}

type providerContextKey struct{}

func getProviderContext(ctx context.Context) *providerContext {
	return ctx.Value(providerContextKey{}).(*providerContext)
}

func resolveToken(resolver core.TokenResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiToken, err := getBearerToken(r.Header)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			provider, pToken, err := resolver.ResolveProviderToken(apiToken)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), providerContextKey{}, &providerContext{
				provider:      provider,
				providerToken: pToken,
				apiToken:      apiToken,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requestLogger(logger zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				if r := recover(); r != nil && r != http.ErrAbortHandler {
					logger.Error().Interface("recover", r).Bytes("stack", debug.Stack()).Msg("incoming_request_panic")
					ww.WriteHeader(http.StatusInternalServerError)
				}
				logger.Info().Fields(map[string]interface{}{
					"remote_addr": r.RemoteAddr,
					"path":        r.URL.Path,
					"proto":       r.Proto,
					"method":      r.Method,
					"user_agent":  r.UserAgent(),
					"status":      http.StatusText(ww.Status()),
					"status_code": ww.Status(),
					"bytes_in":    r.ContentLength,
					"bytes_out":   ww.BytesWritten(),
				}).Msg("incoming_request")
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
