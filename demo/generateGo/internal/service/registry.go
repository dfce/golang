package service

import (
	"generatego/internal/port"

	"go.uber.org/zap"
)

type Registry struct {
	Demo *DemoService
	User *UserService
}

func NewRegistry(
	userRepo port.UserRepository,
	sessions port.SessionStore,
	scripts port.ScriptStore,
	tokens port.TokenService,
	logger *zap.Logger,
) *Registry {
	return &Registry{
		Demo: NewDemoService(scripts, logger),
		User: NewUserService(userRepo, sessions, tokens, logger),
	}
}
