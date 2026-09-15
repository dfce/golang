package repository

import "generatego/internal/platform/datastore"

// 注册表 注册所有 repository
type Registry struct {
	Demo  *DemoRepository
	User  *UserRepository
	Redis *RedisRepository
}

func NewRegistry(dbs datastore.Databases, redis *datastore.RedisClient) *Registry {
	return &Registry{
		Demo:  NewDemoRepository(dbs, redis),
		User:  NewUserRepository(dbs),
		Redis: NewRedisRepository(redis),
	}
}
