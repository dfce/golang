package repository

import (
	"context"
	"errors"
	"testing"

	"generatego/pkg/apperror"
)

func TestRedisRepositoryDisabled(t *testing.T) {
	repo := NewRedisRepository(nil)
	if repo.Enabled() {
		t.Fatal("disabled Redis repository reported enabled")
	}

	if _, _, err := repo.Get(context.Background(), "key"); !errors.Is(err, apperror.ErrRedisDisabled) {
		t.Fatalf("Get() error = %v, want ErrRedisDisabled", err)
	}
	if err := repo.Set(context.Background(), "key", "value", 0); !errors.Is(err, apperror.ErrRedisDisabled) {
		t.Fatalf("Set() error = %v, want ErrRedisDisabled", err)
	}
}
