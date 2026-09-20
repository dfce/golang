package model

// User 对应 PostgreSQL/MYSQL 中的users表
type User struct {
	Base            // 匿名组合，自动继承 ID, CreatedAt, UpdatedAt
	Username string `gorm:"column:username;type:varchar(50);not null;uniqueIndex;comment:用户名"`
	Password string `gorm:"column:password;type:varchar(100);not null;comment:密码"`
	Email    string `gorm:"column:email;type:varchar(100);comment:邮箱"`
	Status   int    `gorm:"column:status;type:smallint;default:1;not null;comment:状态:1启用,2禁用"`
}

// TableName 显示指定表名， 防止GORM 默认加复数 变成 user_s
func (User) TableName() string {
	return "users"
}

func (User) TableComment() string {
	return "用户表"
}
