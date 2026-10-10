package payment

import (
	"context"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 定义聚合所需模块、接口
type UserChecker interface {
	// 在当前事务中锁定指定用户，检查用户是否可进行支付
	LockActiveUser(ctx context.Context, tx *gorm.DB, userID int64) error
}
type WallerChecker interface {
	// 在当前事务中锁定用户指定币种的钱包，并完成扣款
	Debit(ctx context.Context, tx *gorm.DB, userID int64, code string, amount decimal.Decimal) error
}

type Dependencies struct {
	User UserChecker
	// Wallet WallerChecker
}
