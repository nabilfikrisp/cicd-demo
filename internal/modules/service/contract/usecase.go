package contract

import (
	"context"

	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
)

// Usecase describes the service behavior used by delivery adapters.
type Usecase interface {
	Create(context.Context, string, string) (domain.Service, error)
	List(context.Context) ([]domain.Service, error)
}
