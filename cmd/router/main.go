package main

import (
	"net/http"
	"os"

	"magicrouter/core"
	"magicrouter/inmem"
	"magicrouter/openai"
	"magicrouter/server"

	"github.com/rs/zerolog/log"
)

func main() {
	tokenStore := inmem.TokenStore{"test": os.Getenv("OPENAI_API_KEY")}
	services := core.ChatServices{
		"openai": openai.NewChatService(http.DefaultClient),
	}
	svr := server.New(tokenStore, services)
	err := svr.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
