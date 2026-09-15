package service

import (
	"context"
	"errors"
	"generatego/internal/httpserver/middleware"
	"generatego/internal/model"
	"generatego/internal/repository"
	"generatego/pkg/constant"
	"generatego/pkg/jwt"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type UserService struct {
	*BaseService
	repo  *repository.UserRepository
	redis *repository.RedisRepository
}

func NewUserService(repo *repository.UserRepository, redis *repository.RedisRepository, logger *zap.Logger) *UserService {
	return &UserService{&BaseService{Logger: logger}, repo, redis}
}

func (u *UserService) Create(ctx context.Context, body model.CreateUser) (int64, error) {
	var User = &model.User{
		Username: body.Username,
		Password: body.Password,
		Email:    body.Email,
	}

	result := u.repo.CreateUser(ctx, User)
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (s *UserService) Login(ctx context.Context, body model.UserLogin) (string, error) {
	var Login = &model.UserLogin{
		Username: body.Username,
		Password: body.Password,
		Confirm:  body.Confirm,
	}

	if Login.Confirm != Login.Password {
		return "", errors.New("确认密码不一致")
	}

	user, err := s.repo.GetLoginUser(ctx, Login.Username, Login.Password)
	if err != nil {
		return "", err
	}

	if user == nil {
		return "", errors.New("用户不存在")
	}

	var info jwt.UserInfo
	info = jwt.UserInfo{
		Id:   user.Id,
		Name: user.Username,
	}

	token, err := jwt.GenerateToken(info)
	if err != nil {
		return "", err
	}

	// 判定登陆成功后设置 token
	tokenKey := middleware.AuthUserKey + ":" + strconv.FormatInt(info.Id, 10)
	expireDuration := time.Duration(constant.JwtExpire) * time.Second
	s.redis.Set(ctx, tokenKey, token, expireDuration)
	return token, nil
}

func (s *UserService) List(ctx context.Context, query model.GetUser) ([]model.GetUserRes, error) {

	userinfo := s.Userinfo(ctx)
	s.CtxLog(ctx).Debug("current User: ", zap.Any("uinfo", userinfo))

	queryOpt := model.GetUserListOption{
		Id:      query.Id,
		Name:    query.Name,
		Limit:   10,
		Offset:  0,
		OrderBy: "",
	}

	if query.Size != 0 {
		queryOpt.Limit = query.Size
	}
	if query.Page != 0 {
		queryOpt.Offset = (query.Page - 1) * queryOpt.Limit
	}
	// time.Sleep(3 * time.Second)
	return s.repo.GetList(ctx, queryOpt)
}
