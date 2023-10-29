package core

// TokenResolver resolves a provider and a provider token from an api token.
type TokenResolver interface {
	ResolveProviderToken(apiToken string) (provider string, pToken string, err error)
}
