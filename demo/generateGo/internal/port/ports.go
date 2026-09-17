package port

import (
	"context"
	"time"

	"generatego/internal/model"
)

// UserRepository is the persistence contract required by UserService.
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) (int64, error)
	GetLoginUser(ctx context.Context, username string) (*model.User, error)
	GetList(ctx context.Context, opt model.GetUserListOption) ([]model.GetUserRes, error)
}

// SessionStore stores the currently active login session for a user.
type SessionStore interface {
	Set(ctx context.Context, key, value string, expiration time.Duration) error
	Get(ctx context.Context, key string) (value string, found bool, err error)
}

// ScriptStore is the small Redis capability used by the demo service.
type ScriptStore interface {
	ExecScript(ctx context.Context, script string, keys []string, values ...any) (string, error)
}

type Principal struct {
	UserID   int64
	Username string
}

// TokenService hides the concrete token format and signing implementation.
type TokenService interface {
	Issue(ctx context.Context, principal Principal) (token string, ttl time.Duration, err error)
	Verify(ctx context.Context, token string) (Principal, error)
}
