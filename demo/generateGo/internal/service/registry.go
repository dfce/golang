package service

import (
	"generatego/internal/repository"

	"go.uber.org/zap"
)

type Registry struct {
	Demo *DemoService
	User *UserService
}

func NewRegistry(repos *repository.Registry, logger *zap.Logger) *Registry {
	return &Registry{
		Demo: NewDemoService(repos.Demo, repos.Redis, logger),
		User: NewUserService(repos.User, repos.Redis, logger),
	}
}
