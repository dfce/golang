package model

import "github.com/shopspring/decimal"

type Wallet struct {
	Base                   // 匿名组合， 自动继承 Base. ID CreatedAt UpdatedAt
	UserId int64           `gorm:"column:user_id;not null;comment:用户id;index:idx_order_user_currency,priority:1"`
	Code   string          `gorm:"column:code;type:varchar(50);not null;comment:货币代码;idx_order_user_currency,priority:2"`
	Amount decimal.Decimal `gorm:"column:amount;type:numeric(20,2);not null;default:0.00;comment:钱包余额"`
}

func (Wallet) TableName() string {
	return "wallets"
}
func (Wallet) TableComment() string {
	return "用户钱包表"
}
