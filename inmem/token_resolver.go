package inmem

import "errors"

type TokenStore map[string]string

func (s TokenStore) Resolve(apiToken string) (string, error) {
	projectID, ok := s[apiToken]
	if !ok {
		return "", errors.New("project id not found")
	}
	return projectID, nil
}
