package payment

import "github.com/shopspring/decimal"

// 使用扣除用户钱包请求参数
type PayReq struct {
	// 用户id
	UserId int64 `json:"user_id" binding:"required" minimum:"1" example:"1"`
	// 扣除金额
	Amount decimal.Decimal `json:"amount" binding:"required" minimum:"1" example:"1"`
}
