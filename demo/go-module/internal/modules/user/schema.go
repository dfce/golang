package user

// CreateUser 创建用户请求参数。
type CreateUser struct {
	// 用户名，长度 3-50 个字符，且必须唯一。
	Username string `json:"username" binding:"required,min=3,max=50" minLength:"3" maxLength:"50" example:"alice" comment:"用户名"`
	// 登录密码，长度 8-72 个字符。
	Password string `json:"password" binding:"required,min=8,max=72" minLength:"8" maxLength:"72" format:"password" example:"ChangeMe123!" comment:"密码"`
	// 用户邮箱，可选；填写时必须是合法邮箱地址。
	Email string `json:"email" binding:"omitempty,email,max=100" maxLength:"100" format:"email" example:"alice@example.com" comment:"用户邮箱"`
}

// CreateUserRes 创建用户成功后的响应数据。
type CreateUserRes struct {
	// 新创建用户的数据库 ID。
	ID int64 `json:"id" minimum:"1" example:"1"`
}

// UserLogin 用户登录请求参数。
type UserLogin struct {
	// 用户名。
	Username string `json:"username" binding:"required,min=3,max=50" minLength:"3" maxLength:"50" example:"alice"`
	// 登录密码。
	Password string `json:"password" binding:"required,min=8,max=72" minLength:"8" maxLength:"72" format:"password" example:"ChangeMe123!"`
}

// GetUser 用户列表查询参数。
type GetUser struct {
	// 页码，从 1 开始。
	Page int `form:"page" json:"page" binding:"omitempty,gte=1,lte=100000" minimum:"1" default:"1" example:"1" comment:"页码"`
	// 每页数量，范围 1-100。
	Size int `form:"size" json:"size" binding:"omitempty,gte=1,lte=100" minimum:"1" maximum:"100" default:"20" example:"20" comment:"页大小"`
	// 按用户 ID 精确查询。
	ID int64 `form:"id" json:"id" binding:"omitempty,gt=0" minimum:"1" example:"1" comment:"用户 ID"`
	// 按用户名模糊查询。
	Name string `form:"name" json:"name" binding:"omitempty,max=50" maxLength:"50" example:"alice" comment:"用户名"`
}

// GetUserRes 用户列表中的用户数据。
type GetUserRes struct {
	ID       int64  `json:"id" minimum:"1" example:"1"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   int    `json:"status" enums:"1,2" example:"1"`
}

// GetUserListOption 用户列表查询的内部条件。
type GetUserListOption struct {
	ID   int64
	Name string
	// 分页 排序
	OrderBy string // 排序字段、方向. example: id DESC
	Limit   int
	Offset  int
}
