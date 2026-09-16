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

// ====================
// 创建参数
type CreateUser struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	// Email    string `json:"email" binding:"required"`
}

// userLogin
type UserLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 查询用户列表参数
type GetUser struct {
	Page int    `form:"page" json:"page" binding:"" comment:"页码"`
	Size int    `form:"size" json:"size" binding:"" comment:"页大小"`
	Id   int64  `form:"id" json:"id" binding:"" comment:"用户id"`
	Name string `form:"name" json:"name" binding:"" comment:"用户名"`
}

// 用户列表响应结构
type GetUserRes struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   int    `json:"status"`
}

// 查询用户列表条件
type GetUserListOption struct {
	Id   int64
	Name string
	// 分页 排序
	OrderBy string // 排序字段、方向. example: id DESC
	Limit   int
	Offset  int
}
