package memory

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/domain"
)

func TestCreateConcurrentDuplicateAndList(t *testing.T) {
	r := NewServiceRepository()
	var successes atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := r.Create(context.Background(), domain.Service{Name: "billing"})
			if err == nil {
				successes.Add(1)
			} else if !errors.Is(err, domain.ErrServiceAlreadyExists) {
				t.Errorf("create: %v", err)
			}
		}()
	}
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("successful creates = %d", successes.Load())
	}
	if err := r.Create(context.Background(), domain.Service{Name: "accounts"}); err != nil {
		t.Fatal(err)
	}
	services, err := r.List(context.Background())
	if err != nil || len(services) != 2 || services[0].Name != "accounts" || services[1].Name != "billing" {
		t.Fatalf("list = %v, %v", services, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := r.List(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list: %v", err)
	}
}
