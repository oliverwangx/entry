package user

import (
	"encoding/json"
	"log"
	"net/http"
	"shopee-backend-entry-task/client/internal/pkg/storage"
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

	// Map HTTP request to request model
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("unable to decode HTTP request: %v", err)
		return
	}

	// Store request model and map response model
	if err := c.storage.Store(req, &res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to store request: %v", err)
		return
	}

	// Check if the response code is an expected value
	if res.Code != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unexpected storage response: %d", res.Code)
		return
	}

	// Convert response model to HTTP response
	data, err := json.Marshal(res.Data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to marshal response: %v", err)
		return
	}

	// Respond
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	_, _ = w.Write(data)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   res.Data.SessionToken,
		Expires: res.Data.ExpireTime,
	})
}
