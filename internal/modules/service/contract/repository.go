package contract

import (
	"context"

	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
)

// Repository returns services ordered by name and atomically rejects duplicate
// names with domain.ErrServiceAlreadyExists. Names are compared exactly after
// use-case normalization. List returns an empty slice when there are no services.
type Repository interface {
	Create(context.Context, domain.Service) error
	List(context.Context) ([]domain.Service, error)
}
