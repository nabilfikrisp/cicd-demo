package usecase

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/contract"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
)

type ServiceUsecase struct{ repository contract.Repository }

func NewServiceUsecase(repository contract.Repository) *ServiceUsecase {
	return &ServiceUsecase{repository: repository}
}

func (u *ServiceUsecase) Create(ctx context.Context, name, description string) (domain.Service, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 200 {
		return domain.Service{}, fmt.Errorf("%w: name must contain 1 to 200 characters", domain.ErrInvalidService)
	}
	if !utf8.ValidString(description) || utf8.RuneCountInString(description) > 1000 {
		return domain.Service{}, fmt.Errorf("%w: description must contain at most 1000 characters", domain.ErrInvalidService)
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return domain.Service{}, fmt.Errorf("generate service ID: %w", err)
	}
	service := domain.Service{ID: id.String(), Name: name, Description: description}
	if err := u.repository.Create(ctx, service); err != nil {
		return domain.Service{}, err
	}
	return service, nil
}

func (u *ServiceUsecase) List(ctx context.Context) ([]domain.Service, error) {
	return u.repository.List(ctx)
}
