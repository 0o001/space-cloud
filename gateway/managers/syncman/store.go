package syncman

import (
	"context"

	"github.com/spaceuptech/space-cloud/gateway/config"
	"github.com/spaceuptech/space-cloud/gateway/model"
)

// Store abstracts the implementation of letsencrypt storage operations
type Store interface {
	WatchServices(cb func(eventType string, serviceID string, projects model.ScServices)) error
	WatchResources(cb func(eventType, resourceId string, resourceType config.Resource, resource interface{})) error

	Register()

	SetResource(ctx context.Context, resourceID string, resource interface{}) error
	DeleteResource(ctx context.Context, resourceID string) error

	// This function should only be used by delete project endpoint
	DeleteProject(ctx context.Context, projectID string) error

	GetGlobalConfig() (*config.Config, error)
}
