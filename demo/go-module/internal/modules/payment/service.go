package payment

import (
	"context"
	"errors"
	"go-module/internal/platform/datastore"
	"go-module/pkg/apperror"
	"go-module/pkg/base"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentSvc struct {
	*base.BaseService
	repo *Repository
	deps Dependencies
}

func NewService(repo *Repository, logger *zap.Logger, redis *datastore.RedisSvc, deps Dependencies) *PaymentSvc {
	return &PaymentSvc{
		BaseService: &base.BaseService{Logger: logger},
		repo:        repo,
		deps:        deps,
	}
}

// ==================================== 业务逻辑 ====================================
func (s *PaymentSvc) Pay(ctx context.Context, req PayReq) (int64, error) {
	// if s.deps.User == nil || s.deps.Wallet == nil{
	if s.deps.User == nil {
		return 0, apperror.Internal("支付依赖未配置")
	}

	var orderId int64

	err := s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		// 统一锁顺序：先锁用户，再锁用户钱包
		if err := s.deps.User.LockActiveUser(
			ctx, tx, req.UserId,
		); err != nil {
			return err
		}
		return errors.New("test rollback")
		return nil
	})

	return orderId, err
}
