package user

import (
	"context"
	"errors"
	model "go-module/internal/models"
	"go-module/pkg/apperror"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

const (
	postgresUniqueViolationCode = "23505"
	usernameUniqueConstraint    = "idx_users_username"
)

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db}
}

func duplicateKeyError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolationCode {
		message := "数据已存在"
		if pgErr.ConstraintName == usernameUniqueConstraint {
			message = "用户名已存在"
		}
		return apperror.Wrap(err, http.StatusConflict, message)
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperror.Wrap(
			err,
			http.StatusConflict,
			"数据已存在",
		)
	}
	return err
}

// ==================================== 业务逻辑 ====================================
func (repo *Repository) Create(ctx context.Context, user *model.User) (int64, error) {
	res := repo.db.WithContext(ctx).Create(user)
	if res.Error != nil {
		// 数据库唯一约束是并发场景下的最终保障。
		return 0, duplicateKeyError(res.Error)
	}
	return user.Id, nil
}

func (r *Repository) GetByUserame(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		// Select("*").
		Select("id", "username", "password").
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
