package core

type TokenResolver interface {
	Resolve(apiToken string) (projectID string, err error)
}
