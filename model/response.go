package model

import "time"

type Response interface {
	GetRequestType() string
	GetStatusCode() int
}

type BaseResponse struct {
	Code        int    `json:"code"`
	RequestType string `json:"request_type"`
}

type LoginData struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	NickName     string `json:"nick_name"`
	AvatarPath   string `json:"avatar_path"`
	SessionToken string `json:"session_token"`
}

type NickNameData struct {
	NickName string `json:"nick_name"`
}

type AvatarData struct {
	AvatarPath string `json:"avatar_path"`
}

type LoginResponse struct {
	Code         int       `json:"code"`
	RequestType  string    `json:"request_type"`
	Data         LoginData `json:"data"`
	SessionToken string    `json:"sessionToken"`
	ExpireTime   time.Time `json:"Expires"`
}

type NickNameResponse struct {
	Code        int          `json:"code"`
	RequestType string       `json:"request_type"`
	Data        NickNameData `json:"data"`
}

type AvatarResponse struct {
	Code        int        `json:"code"`
	RequestType string     `json:"request_tzype"`
	Data        AvatarData `json:"data"`
}

func (r BaseResponse) GetRequestType() string {
	return r.RequestType
}

func (r BaseResponse) GetStatusCode() int {
	return r.Code
}
