package redisCache

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"oliver/entry/config"
	"oliver/entry/model"
	"oliver/entry/utils/logger"
	"time"
)

type CacheStore struct {
	rds *redis.Client
}

func (c *CacheStore) Init(serverConfig map[string]string) {
	c.rds = redis.NewClient(&redis.Options{
		Addr:     serverConfig[config.RedisHost] + serverConfig[config.RedisPort],
		Password: "",
		DB:       0,
		DialTimeout: 1 * time.Second,
	})
}

func (c *CacheStore) GetUserByUsername(ctx context.Context, username string) (user *model.User, err error) {
	var val string
	user = new(model.User)
	if val, err = c.rds.Get(ctx, username).Result(); err != nil {
		return
	}

	err = json.Unmarshal([]byte(val), user)
	return
}

func (c *CacheStore) SetUser(ctx context.Context, username string, user *model.User) (err error) {
	var data []byte
	if data, err = json.Marshal(*user); err != nil {
		logger.Error.Println("user data Marshal problem is ", err)
		return
	}
	err = c.rds.Set(ctx, username, data, 0).Err()
	if err != nil {
		logger.Error.Println("cache fails store information", err)
	}
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
