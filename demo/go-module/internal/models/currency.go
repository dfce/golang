package model

type Currency struct {
	Base        // 匿名组合， 自动继承 Base. ID CreatedAt UpdatedAt
	Code string `gorm:"column:code;type:varchar(50);not null;comment:货币代码;index:idx_currency_code,uinque,comment:全球货币通用代码"`
	Name string `gorm:"column:name;type:varchar(50);not null;comment:币种名称"`
}

func (Currency) TableName() string {
	return "currency"
}
func (Currency) TableComment() string {
	return "货币表"
}
