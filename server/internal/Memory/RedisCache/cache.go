package redisCache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"shopee-backend-entry-task/model"
)

type CacheStore struct {
	rds *redis.Client
}

func (c *CacheStore) Init() {
	c.rds = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

func (c *CacheStore) GetUserByUsername(ctx context.Context, username string) (user *model.User, err error) {
	var val string
	if val, err = c.rds.Get(ctx, username).Result(); err != nil {
		return
	}
	err = json.Unmarshal([]byte(val), user)
	return
}

func (c *CacheStore) SetUser(ctx context.Context, username string, user *model.User) (err error) {
	var data []byte
	if data, err = json.Marshal(*user); err != nil {
		return
	}
	err = c.rds.Set(ctx, username, data, 0).Err()
	return
}

func (c *CacheStore) DeleteUser(ctx context.Context, username string) (err error) {
	err = c.rds.Del(ctx, username).Err()
	return
}

func (c *CacheStore) SetUserSession(ctx context.Context, username string, token string) (err error) {
	err = c.rds.Set(ctx, token, username, 0).Err()
	return
}

func (c *CacheStore) GetUserSession(ctx context.Context, token string) (username string, err error) {
	username, err = c.rds.Get(ctx, token).Result()
	return
}

func (c *CacheStore) DeleteUserSession(ctx context.Context, username string) (err error) {
	err = c.rds.Del(ctx, "Session/"+username).Err()
	return
}
