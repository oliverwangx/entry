package Memory

import (
	"context"
	"shopee-backend-entry-task/model"
	"shopee-backend-entry-task/server/internal/Memory/MySQLDB"
	"shopee-backend-entry-task/server/internal/Memory/RedisCache"
	logger2 "shopee-backend-entry-task/utils/logger"
)

type DataStore struct {
	DB    *sqlDB.DBStore
	Cache *redisCache.CacheStore
}

func (d *DataStore) Init(serverConfig map[string]string) (err error) {

	d.Cache = new(redisCache.CacheStore)
	d.DB = new(sqlDB.DBStore)
	d.Cache.Init(serverConfig)
	err = d.DB.Init()
	return
}

func (d *DataStore) GetUserByUsername(ctx context.Context, username string) (user *model.User, err error) {
	// fetch user information from cache
	if user, err = d.Cache.GetUserByUsername(ctx, username); user != nil && err == nil {
		logger2.Info.Println("Get User in Cache")
		return
	}
	// fetch from sql database
	if user, err = d.DB.GetUserByUsername(username); err != nil {
		logger2.Error.Println("DataBase Fetch Data Error", err)
		return
	}
	// fmt.Println(user)
	// add user to cache
	err = d.Cache.SetUser(ctx, username, user)
	return
}

func (d *DataStore) UpdateUserAvatar(ctx context.Context, userName string, url string) (err error) {
	if err = d.DB.UpdateUserAvatar(userName, url); err != nil {
		return
	}
	// clear cache
	err = d.Cache.DeleteUser(ctx, userName)
	// err = d.Cache.SetUser(ctx, user.Username, user)
	return
}

func (d *DataStore) UpdateUserNickname(ctx context.Context, userName string, nickName string) (err error) {
	if err = d.DB.UpdateUserNickname(userName, nickName); err != nil {
		return
	}
	err = d.Cache.DeleteUser(ctx, userName)
	// err = d.Cache.SetUser(ctx, user.Username, user)
	return
}

func (d *DataStore) SetUserSession(ctx context.Context, username string, token string) (err error) {
	// if err = d.db.SetUserSession(username, token); err != nil {
	//	return
	//}
	err = d.Cache.SetUserSession(ctx, username, token)
	return
}

func (d *DataStore) GetUserSession(ctx context.Context, username string) (token string, err error) {
	token, err = d.Cache.GetUserSession(ctx, username)
	return
}

func (d *DataStore) DeleteUserSession(ctx context.Context, username string) (err error) {
	err = d.Cache.DeleteUserSession(ctx, username)
	return
}
