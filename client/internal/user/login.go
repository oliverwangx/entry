package user

import (
	"encoding/json"
	"net/http"
	"shopee-backend-entry-task/client/internal/pkg/storage"
	"shopee-backend-entry-task/logger"
	"shopee-backend-entry-task/model"
)

type Login struct {
	storage storage.Storage
}

// NewLogIn Constructor
func NewLogIn(str storage.Storage) Login {
	return Login{
		storage: str,
	}
}

// Handle POST /api/v1/users
func (c Login) Handle(w http.ResponseWriter, r *http.Request) {
	var (
		req model.LogInParams
		res model.LoginResponse
	)
	req.RequestType = "login"
	// Map HTTP request to request model
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Println("unable to decode HTTP request: %v", err)
		return
	}

	if req.RequestType == "" {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Println("Can not recognize the request type")
		return
	}
	// Store request model and map response model
	if err := c.storage.Store(req, &res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Println("unable to store request: %v", err)
		return
	}

	// Check if the response code is an expected value
	if res.Code != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Println("unexpected storage response: %d", res.Code)
		return
	}

	// Convert response model to HTTP response
	data, err := json.Marshal(res.Data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Println("unable to marshal response: %v", err)
		return
	}

	// Respond
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Header().Add("Access-Control-Allow-Origin ", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers ", "Access-Control-Allow-Headers ")

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   res.SessionToken,
		Expires: res.ExpireTime,
	})
	_, _ = w.Write(data)
}
