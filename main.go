package main

import (
	"net/http"
	"os"

	"magicrouter/inmem"
	"magicrouter/providers"
	"magicrouter/server"
)

func main() {
	tokenStore := inmem.TokenStore{"test": os.Getenv("OPENAI_API_KEY")}
	services := map[string]server.ChatService{
		"openai": providers.NewOpenAIChatService(http.DefaultClient),
	}
	svr := server.New(tokenStore, services)
	err := svr.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
