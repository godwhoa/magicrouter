package core

type ProjectConfig struct {
	ID     string  `json:"id"`
	Routes []Route `json:"routes"`
}

type ProjectStore interface {
	GetConfig(projectID string) (*ProjectConfig, error)
}
