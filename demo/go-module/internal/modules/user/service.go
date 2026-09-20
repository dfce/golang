package user

import (
	"context"
	model "go-module/internal/models"
	"go-module/pkg/base"

	"go.uber.org/zap"
)

type UserSvc struct {
	*base.BaseService
	repo *Repository
}

func NewService(userRepo *Repository, logger *zap.Logger) *UserSvc {
	return &UserSvc{
		BaseService: &base.BaseService{Logger: logger},
		repo:        userRepo,
	}
}

// ==================================== 业务逻辑 ====================================
func (s *UserSvc) Create(ctx context.Context, body CreateUser) (int64, error) {
	user := &model.User{
		Username: body.Username,
		Password: body.Password,
		Email:    body.Email,
		Status:   1,
	}

	res, err := s.repo.Create(ctx, user)
	if err != nil {
		return 0, err
	}

	s.CtxLog(ctx).Info("创建用户成功", zap.Any("user", res))

	return res, nil
}
