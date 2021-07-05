package model

type BasicParams struct {
	Username    string `json:"username"`
	RequestType string `json:"request_type"`
}

type LogInParams struct {
	BasicParams
	Password string `json:"password"`
}
