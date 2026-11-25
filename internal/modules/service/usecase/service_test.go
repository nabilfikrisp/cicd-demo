package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/repository/memory"
)

func TestCreateValidation(t *testing.T) {
	u := NewServiceUsecase(memory.NewServiceRepository())
	for _, tc := range []struct{ name, description string }{{" ", ""}, {strings.Repeat("a", 201), ""}, {"valid", strings.Repeat("a", 1001)}} {
		if _, err := u.Create(context.Background(), tc.name, tc.description); !errors.Is(err, domain.ErrInvalidService) {
			t.Errorf("expected validation error: %v", err)
		}
	}
	service, err := u.Create(context.Background(), " billing ", " invoices ")
	if err != nil || service.Name != "billing" || service.Description != "invoices" {
		t.Fatalf("create: %+v, %v", service, err)
	}
	if _, err := uuid.Parse(service.ID); err != nil {
		t.Fatalf("invalid ID: %v", err)
	}
	if _, err := u.Create(context.Background(), "billing", ""); !errors.Is(err, domain.ErrServiceAlreadyExists) {
		t.Fatalf("duplicate: %v", err)
	}
}
