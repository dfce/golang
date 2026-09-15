package repository

import (
	"context"
	"errors"
	"generatego/internal/model"
	"generatego/internal/platform/datastore"

	"gorm.io/gorm"
)

type UserRepository struct {
	dbs datastore.Databases
}

func NewUserRepository(dbs datastore.Databases) *UserRepository {
	return &UserRepository{dbs}
}

func (u *UserRepository) CreateUser(ctx context.Context, user *model.User) *gorm.DB {
	return u.dbs["primary"].WithContext(ctx).Create(user)
}

func (u *UserRepository) GetLoginUser(ctx context.Context, username, Password string) (*model.User, error) {
	var user model.User

	err := u.dbs["primary"].
		WithContext(ctx).
		Select("Id", "Username").
		Where("Username = ? and Password = ?", username, Password).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetList(ctx context.Context, opt model.GetUserListOption) ([]model.GetUserRes, error) {
	var list []model.GetUserRes

	query := r.dbs["primary"].WithContext(ctx).Model(&model.User{})

	if opt.Id != 0 {
		query = query.Where("id = ?", opt.Id)
	}
	if opt.Name != "" {
		query = query.Where("username LIKE ?", opt.Name+"%")
	}

	// 组装 where
	if opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}
	if opt.Offset > 0 {
		query = query.Offset(opt.Offset)
	}
	if opt.OrderBy != "" {
		query = query.Order(opt.OrderBy)
	} else {
		query = query.Order("id DESC")
	}

	err := query.Find(&list).Error
	return list, err
}
