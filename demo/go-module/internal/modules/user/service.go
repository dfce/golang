package user

import (
	"context"
	"encoding/json"
	model "go-module/internal/models"
	"go-module/internal/platform/datastore"
	"go-module/pkg/apperror"
	"go-module/pkg/base"
	"go-module/pkg/constant"
	"go-module/pkg/jwt"
	"net/http"
	"strconv"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserSvc struct {
	*base.BaseService
	repo         *Repository
	tokenService *jwt.Service
	redisSvc     *datastore.RedisSvc
}

func NewService(userRepo *Repository, logger *zap.Logger, tokenService *jwt.Service, redis *datastore.RedisSvc) *UserSvc {
	return &UserSvc{
		BaseService:  &base.BaseService{Logger: logger},
		repo:         userRepo,
		tokenService: tokenService,
		redisSvc:     redis,
	}
}

// ==================================== 业务逻辑 ====================================
func (s *UserSvc) Create(ctx context.Context, body CreateUser) (int64, error) {
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

	res, err := s.repo.Create(ctx, user)
	if err != nil {
		return 0, err
	}

	s.CtxLog(ctx).Info("创建用户成功", zap.Any("user", res))

	return res, nil
}

func (s *UserSvc) Login(ctx context.Context, body UserLogin) (string, error) {

	user, err := s.repo.GetByUserame(ctx, body.Username)
	if err != nil {
		return "", err
	}

	// data, _ := json.Marshal(user)
	data, _ := json.MarshalIndent(user, "", " ")

	s.CtxLog(ctx).Debug("", zap.String("user", string(data)))
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)) != nil {
		return "", apperror.Unauthorized("用户名或密码错误")
	}

	info := jwt.UserInfo{
		Id:   user.Id,
		Name: user.Username,
	}

	token, err := s.tokenService.GenerateToken(info)
	if err != nil {
		return "", apperror.Wrap(err, http.StatusInternalServerError, "生成认证令牌失败")
	}

	tokenKey := constant.AuthUserKey + strconv.FormatInt(info.Id, 10)
	expireDuration := s.tokenService.Expiry()

	if s.redisSvc != nil {
		if err := s.redisSvc.Set(ctx, tokenKey, token, expireDuration).Err(); err != nil {
			return "", apperror.Wrap(err, http.StatusServiceUnavailable, "认证服务暂不可用")
		}
	}

	val, ok, err := s.redisSvc.GetString(ctx, tokenKey)
	if err != nil {
		s.CtxLog(ctx).Error("获取缓存失败")
		return "", nil
	}

	if !ok {
		s.CtxLog(ctx).Info("缓存数据不存在")
	}

	s.CtxLog(ctx).Info("获取缓存成功", zap.Any("val", val))

	return token, nil
}

func (s *UserSvc) List(ctx context.Context) {

}
