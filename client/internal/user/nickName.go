package user

import (
	"encoding/json"
	"net/http"
	"oliver/entry/client/internal/pkg/storage"
	"oliver/entry/model"
	"oliver/entry/utils/logger"
)

type NickName struct {
	storage storage.Storage
}

// NewNickName Constructor
func NewNickName(str storage.Storage) NickName {
	return NickName{
		storage: str,
	}
}

// Handle POST /api/v1/nicknames
func (n NickName) Handle(w http.ResponseWriter, r *http.Request) {
	var (
		req model.NickNameParams
		res model.NickNameResponse
	)
	// fmt.Println("got in")
	req.RequestType = "update_nickname"
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Println("unable to decode HTTP request: %v", err)
		return
	}

	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			logger.Error.Println("Request StatusUnauthorized")
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		logger.Error.Println("StatusBadRequest")
		return
	}
	sessionToken := c.Value
	req.SessionToken = sessionToken

	if err := n.storage.Store(req, &res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Println("unable to store request: %v", err)
		return
	}

	if res.Code != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Println("unexpected storage response: %d", res.Code)
		return
	}

	data, err := json.Marshal(res.Data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error.Println("unable to marshal response: %v", err)
		return
	}

	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Header().Add("Access-Control-Allow-Origin ", "http://127.0.0.1:5500")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers ", "Access-Control-Allow-Headers ")

	_, _ = w.Write(data)
}
