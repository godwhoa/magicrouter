package inmem

import (
	"errors"

	"magicrouter/core"
)

type ProjectStore map[string]*core.ProjectConfig

func (s ProjectStore) GetConfig(projectID string) (*core.ProjectConfig, error) {
	project, ok := s[projectID]
	if !ok {
		return nil, errors.New("project not found")
	}
	return project, nil
}
