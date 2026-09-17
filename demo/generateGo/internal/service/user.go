package service

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"generatego/internal/model"
	"generatego/internal/port"
	"generatego/pkg/apperror"
	"generatego/pkg/constant"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	*BaseService
	repo     port.UserRepository
	sessions port.SessionStore
	tokens   port.TokenService
}

func NewUserService(repo port.UserRepository, sessions port.SessionStore, tokens port.TokenService, logger *zap.Logger) *UserService {
	return &UserService{
		BaseService: &BaseService{Logger: logger},
		repo:        repo,
		sessions:    sessions,
		tokens:      tokens,
	}
}

func (u *UserService) Create(ctx context.Context, body model.CreateUser) (int64, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, apperror.Wrap(err, http.StatusInternalServerError, "密码处理失败")
	}

	user := &model.User{
		Username: body.Username,
		Password: string(passwordHash),
		Email:    body.Email,
		Status:   1,
	}

	rows, err := u.repo.CreateUser(ctx, user)
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func (u *UserService) Login(ctx context.Context, body model.UserLogin) (string, error) {
	if u.sessions == nil {
		return "", apperror.ErrRedisDisabled
	}
	if u.tokens == nil {
		return "", apperror.ServiceUnavailable("JWT 认证服务未配置")
	}

	user, err := u.repo.GetLoginUser(ctx, body.Username)
	if err != nil {
		return "", err
	}

	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) != nil {
		return "", apperror.Unauthorized("用户名或密码错误")
	}

	principal := port.Principal{
		UserID:   user.Id,
		Username: user.Username,
	}
	token, expireDuration, err := u.tokens.Issue(ctx, principal)
	if err != nil {
		return "", apperror.Wrap(err, http.StatusInternalServerError, "生成认证令牌失败")
	}

	tokenKey := constant.AuthUserKey + ":" + strconv.FormatInt(principal.UserID, 10)
	if err := u.sessions.Set(ctx, tokenKey, token, expireDuration); err != nil {
		if errors.Is(err, apperror.ErrRedisDisabled) {
			return "", err
		}
		return "", apperror.Wrap(err, http.StatusServiceUnavailable, "认证服务暂不可用")
	}

	return token, nil
}

func (u *UserService) List(ctx context.Context, query model.GetUser) ([]model.GetUserRes, error) {
	if userInfo, ok := u.Userinfo(ctx); ok {
		u.CtxLog(ctx).Debug("current user", zap.Int64("id", userInfo.UserID))
	}

	queryOpt := model.GetUserListOption{
		Id:      query.Id,
		Name:    query.Name,
		Limit:   10,
		Offset:  0,
		OrderBy: "",
	}

	if query.Size > 0 {
		queryOpt.Limit = query.Size
	}
	if query.Page > 0 {
		queryOpt.Offset = (query.Page - 1) * queryOpt.Limit
	}

	return u.repo.GetList(ctx, queryOpt)
}
