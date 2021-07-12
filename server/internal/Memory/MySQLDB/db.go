package sqlDB

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"oliver/entry/model"
	"time"
)

type DBStore struct {
	db *sql.DB
}

func (d *DBStore) Init() (err error) {
	d.db, err = sql.Open("mysql", "root:LEle950822@tcp(127.0.0.1:3306)/log_in_system?timeout=3s")
	d.db.SetMaxIdleConns(1000)
	d.db.SetMaxOpenConns(1000)
	d.db.SetConnMaxLifetime(300 * time.Second)
	return
}

func (d *DBStore) GetUserByUsername(username string) (user *model.User, err error) {
	user = new(model.User)
	err = d.db.QueryRow("SELECT username, password, avatar, nickname FROM User WHERE username = ?", username).Scan(&user.Username, &user.Password, &user.Avatar, &user.Nickname)
	//logger2.Info.Println("sql: get user info: ", user)
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

