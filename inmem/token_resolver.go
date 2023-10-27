package inmem

import "errors"

type TokenStore map[string]string

func (s TokenStore) ResolveProviderToken(apiToken string) (string, string, error) {
	token, ok := s[apiToken]
	if !ok {
		return "", "", errors.New("provider token not found")
	}
	return "openai", token, nil
}
