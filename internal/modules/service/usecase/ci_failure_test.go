package usecase

import "testing"

func TestIntentionalCIFailure(t *testing.T) {
	t.Fatal("intentional failure to verify the required CI test check")
}
