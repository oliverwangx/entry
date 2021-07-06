package model

type BasicParams struct {
	Username    string `json:"username"`
	RequestType string `json:"request_type"`
}

type LogInParams struct {
	Username    string `json:"username"`
	RequestType string `json:"request_type"`
	Password    string `json:"password"`
}

type NickNameParams struct {
	SessionToken string `json:"username"`
	RequestType  string `json:"request_type"`
	NickName     string `json:"nick_name"`
}

type AvatarParams struct {
	SessionToken string `json:"username"`
	RequestType  string `json:"request_type"`
	AvatarPath   string `json:"avatar_path"`
}
