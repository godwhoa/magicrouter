package main

import (
	"os"

	"magicrouter/chat"
	"magicrouter/inmem"
	"magicrouter/server"

	"github.com/rs/zerolog/log"
)

func main() {
	tokenStore := inmem.TokenStore{"test": os.Getenv("OPENAI_API_KEY")}
	services := chat.NewServiceMap()
	svr := server.New(tokenStore, services)
	err := svr.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start server")
	}
}
