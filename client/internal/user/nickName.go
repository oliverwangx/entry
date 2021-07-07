package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"shopee-backend-entry-task/client/internal/pkg/storage"
	"shopee-backend-entry-task/model"
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
	fmt.Println("got in")
	req.RequestType = "update_nickname"
	// Map HTTP request to request model
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("unable to decode HTTP request: %v", err)
		return
	}
	fmt.Println("got in1")
	fmt.Println(r.Cookie)
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("StatusUnauthorized")
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("StatusBadRequest")
		return
	}
	sessionToken := c.Value
	req.SessionToken = sessionToken
	fmt.Println("got in2")
	// Store request model and map response model
	if err := n.storage.Store(req, &res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to store request: %v", err)
		return
	}
	fmt.Println("got in3")
	// Check if the response code is an expected value
	if res.Code != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unexpected storage response: %d", res.Code)
		return
	}
	fmt.Println("got in4")
	// Convert response model to HTTP response
	data, err := json.Marshal(res.Data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("unable to marshal response: %v", err)
		return
	}
	fmt.Println("get data5", data)
	// Respond
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.Header().Add("Access-Control-Allow-Origin ", "http://127.0.0.1:5500")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers ", "Access-Control-Allow-Headers ")

	_, _ = w.Write(data)
}
