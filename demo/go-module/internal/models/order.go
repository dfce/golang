package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Order struct {
	Base                              // 匿名组合， 自动继承 Base. ID CreatedAt UpdatedAt
	UserId            int64           `gorm:"column:user_id;not null;comment:用户id;index:idx_order_user_currency,priority:1"`
	Code              string          `gorm:"column:code;type:varchar(50);not null;comment:货币代码;index:idx_order_user_currency,priority:2"`
	Amount            decimal.Decimal `gorm:"column:amount;type:numeric(20,2);not null;comment:金额"`
	TripartiteOrderNo string          `gorm:"column:tripartite_order_no;type:varchar(150);not null;comment:三方订单号;index:idx_order_no,unique"`
	PayTime           time.Time       `gorm:"column:paid_at;not null;comment:支付时间"`
}

func (Order) TableName() string {
	return "orders"
}
func (Order) TableComment() string {
	return "订单表"
}
