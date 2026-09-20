package model

import "time"

type Base struct {
	Id        int64     `gorm:"primaryKey;autoIncrement;comment:主键ID"`
	CreatedAt time.Time `gorm:"column:created_at;not null;comment:创建时间"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;comment:更新时间"`
}

var Models = []any{
	&User{},
}
