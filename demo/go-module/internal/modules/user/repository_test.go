package user

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"go-module/pkg/apperror"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func TestPostgresDuplicateUsernameIsConflict(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:           postgresUniqueViolationCode,
		ConstraintName: usernameUniqueConstraint,
	}
	err := duplicateKeyError(fmt.Errorf("create user: %w", pgErr))
	if err == nil {
		t.Fatal("expected duplicate key error to be mapped")
	}

	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected apperror.Error, got %T", err)
	}
	if appErr.Status != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, appErr.Status)
	}
	if appErr.Message != "用户名已存在" {
		t.Fatalf("unexpected message: %q", appErr.Message)
	}
}

func TestGenericDuplicateKeyIsConflict(t *testing.T) {
	err := duplicateKeyError(gorm.ErrDuplicatedKey)

	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		t.Fatalf("expected apperror.Error, got %T", err)
	}
	if appErr.Status != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, appErr.Status)
	}
	if appErr.Message != "数据已存在" {
		t.Fatalf("unexpected message: %q", appErr.Message)
	}
}

func TestNonDuplicateErrorIsUnchanged(t *testing.T) {
	want := errors.New("database unavailable")
	if got := duplicateKeyError(want); !errors.Is(got, want) {
		t.Fatalf("expected original error, got %v", got)
	}
}
