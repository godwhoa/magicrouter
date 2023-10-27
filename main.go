package main

import (
	"os"

	"magicrouter/chat"
	"magicrouter/inmem"
	"magicrouter/server"
)

func main() {
	tokenStore := inmem.TokenStore{"test": os.Getenv("OPENAI_API_KEY")}
	services := chat.NewServiceMap()
	svr := server.New(tokenStore, services)
	err := svr.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
