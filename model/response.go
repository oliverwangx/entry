package model

import "time"

type Response interface {
	GetRequestType() string
	GetStatusCode() int
}

type BaseResponse struct {
	Code        int `json:"code"`
	RequestType string
}

type LoginResponse struct {
	BaseResponse
	Data struct {
		ID           string    `json:"id"`
		CreatedAt    string    `json:"created_at"`
		SessionToken string    `json:"sessionToken"`
		ExpireTime   time.Time `json:"Expires"`
	} `json:"data"`
}

func (r BaseResponse) GetRequestType() string {
	return r.RequestType
}

func (r BaseResponse) GetStatusCode() int {
	return r.Code
}
