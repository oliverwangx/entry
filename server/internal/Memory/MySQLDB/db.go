package sqlDB

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"shopee-backend-entry-task/model"
	logger2 "shopee-backend-entry-task/utils/logger"
	"time"
)

type DBStore struct {
	db *sql.DB
}

func (d *DBStore) Init() (err error) {
	d.db, err = sql.Open("mysql", "root:LEle950822@tcp(127.0.0.1:3306)/log_in_system")
	d.db.SetMaxIdleConns(1000)
	d.db.SetMaxOpenConns(1000)
	d.db.SetConnMaxLifetime(300 * time.Second)
	return
}

func (d *DBStore) GetUserByUsername(username string) (user *model.User, err error) {
	user = new(model.User)
	err = d.db.QueryRow("SELECT username, password, avatar, nickname FROM User WHERE username = ?", username).Scan(&user.Username, &user.Password, &user.Avatar, &user.Nickname)
	logger2.Info.Println("sql: get user info: ", user)
	return
}

func (d *DBStore) UpdateUserAvatar(username string, url string) (err error) {
	_, err = d.db.Exec("UPDATE User Set avatar = ? WHERE username = ?", url, username)
	return
}

func (d *DBStore) UpdateUserNickname(username string, nickname string) (err error) {
	_, err = d.db.Exec("UPDATE User Set nickname = ? WHERE username = ?", nickname, username)
	return
}

//
//func (d *DBStore) SetUserSession(username string, token string) (err error) {
//	_, err = d.db.Exec("INSERT INTO Session (username, session) VALUES (?, ?)", username, token)
//	return
//}
//
//func (d *DBStore) GetUserSession(username string) (token string, err error) {
//	err = d.db.QueryRow("SELECT session FROM Session WHERE username = ?", username).Scan(&token)
//	return
//}
//
//func (d *DBStore) DeleteUserSession(username string) (err error) {
//	_, err = d.db.Exec("DELETE FROM Session Where username = ?", username)
//	return
//}
