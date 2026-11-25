package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
)

type ServiceRepository struct {
	mu       sync.RWMutex
	services map[string]domain.Service
}

func NewServiceRepository() *ServiceRepository {
	return &ServiceRepository{services: make(map[string]domain.Service)}
}

func (r *ServiceRepository) Create(ctx context.Context, service domain.Service) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, exists := r.services[service.Name]; exists {
		return domain.ErrServiceAlreadyExists
	}
	r.services[service.Name] = service
	return nil
}

func (r *ServiceRepository) List(ctx context.Context) ([]domain.Service, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	services := make([]domain.Service, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}
	sort.Slice(services, func(i, j int) bool { return services[i].Name < services[j].Name })
	return services, nil
}
