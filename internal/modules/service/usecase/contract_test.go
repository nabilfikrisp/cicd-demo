package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
)

type failingRepository struct{ err error }

func (r failingRepository) Create(context.Context, domain.Service) error   { return r.err }
func (r failingRepository) List(context.Context) ([]domain.Service, error) { return nil, r.err }

func TestRepositoryErrorsPropagate(t *testing.T) {
	want := errors.New("repository unavailable")
	u := NewServiceUsecase(failingRepository{err: want})
	if _, err := u.Create(context.Background(), "billing", ""); !errors.Is(err, want) {
		t.Fatalf("create error = %v", err)
	}
	if _, err := u.List(context.Background()); !errors.Is(err, want) {
		t.Fatalf("list error = %v", err)
	}
}
