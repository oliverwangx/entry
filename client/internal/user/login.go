package user

import (
	"encoding/json"
	"net/http"
	"shopee-backend-entry-task/client/internal/pkg/storage"
	"shopee-backend-entry-task/model"
	logger2 "shopee-backend-entry-task/utils/logger"
)

type Login struct {
	storage storage.Storage
}

var count int

func init() {
	count = 0
}

// NewLogIn Constructor
func NewLogIn(str storage.Storage) Login {
	return Login{
		storage: str,
	}
}

// Handle POST /api/v1/users
func (c Login) Handle(w http.ResponseWriter, r *http.Request) {
	//count += 1
	//logger2.Info.Println("计数", count)
	var (
		req model.LogInParams
		res model.LoginResponse
	)
	req.RequestType = "login"
	// Map HTTP request to request model
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger2.Error.Println("unable to decode HTTP request: %v", err)
		return
	}

	if req.RequestType == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger2.Error.Println("Can not recognize the request type")
		return
	}
	// Store request model and map response model
	// start := time.Now() // 获取当前时间
	if err := c.storage.Store(req, &res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger2.Error.Println("unable to store request: %v", err)
		return
	}
	// elapsed := time.Since(start)
	// count += 1
	//fmt.Println("该函数执行完成耗时：", elapsed, "计数", count)

	// Check if the response code is an expected value
	if res.Code != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		logger2.Error.Println("unexpected storage response: ", res.Code)
		return
	}

	// Convert response model to HTTP response
	data, err := json.Marshal(res.Data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger2.Error.Println("unable to marshal response: %v", err)
		return
	}

	// Respond
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	//w.Header().Add("Access-Control-Allow-Origin ", "*")
	//w.Header().Add("Access-Control-Allow-Credentials", "true")
	//w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	//w.Header().Add("Access-Control-Allow-Headers ", "Access-Control-Allow-Headers ")

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   res.SessionToken,
		Expires: res.ExpireTime,
	})
	_, _ = w.Write(data)
}
