package user

import (
	"encoding/json"
	"net/http"
	"oliver/entry/client/internal/pkg/storage"
	"oliver/entry/model"
	"oliver/entry/utils/logger"
)

type Avatar struct {
	storage storage.Storage
}

// NewAvatar Constructor
func NewAvatar(str storage.Storage) Avatar {
	return Avatar{
		storage: str,
	}
}

// Handle POST /api/v1/avator
func (a Avatar) Handle(w http.ResponseWriter, r *http.Request) {
	var (
		req model.AvatarParams
		res model.AvatarResponse
	)

	req.RequestType = "update_avatar"
	// Map HTTP request to request model

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Println("unable to decode HTTP request: %v", err)
		return
	}

	// Get Cookie/Session_token
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value
	req.SessionToken = sessionToken

	// Store request model and map response model
	if err := a.storage.Store(req, &res); err != nil {
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
	w.Header().Add("Access-Control-Allow-Origin ", "http://127.0.0.1:5500")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers ", "Access-Control-Allow-Headers ")
	_, _ = w.Write(data)
}
