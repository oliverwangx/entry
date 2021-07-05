package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/satori/go.uuid"
	"io"
	"log"
	"net"
	"net/http"
	"shopee-backend-entry-task/model"
	"shopee-backend-entry-task/requestType"
	"strings"
	"time"
)

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
	"wxy":   "950822",
}

var cache redis.Conn

func main() {
	initCache()
	listener, err := net.Listen("tcp", "0.0.0.0:8989")
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	for {
		con, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go receiveClientRequest(con)
	}
}

func initCache() {
	conn, err := redis.DialURL("redis://localhost")
	if err != nil {
		panic(err)
	}
	cache = conn
}

func receiveClientRequest(con net.Conn) {
	defer con.Close()
	var response []byte
	clientReader := bufio.NewReader(con)
	for {
		clientRequest, err := clientReader.ReadString('\n')
		switch err {
		case nil:
			clientRequest := strings.TrimSpace(clientRequest)
			if response, err = handleRequest([]byte(clientRequest)); err != nil {
				fmt.Println("handle request error: " + err.Error())
			}
			fmt.Println(response)
			if clientRequest == ":QUIT" {
				log.Println("client requested server to close the connection so closing")
				return
			}
		case io.EOF:
			log.Println("client closed the connection by terminating the process")
			return
		default:
			log.Printf("error: %v\n", err)
			return
		}

		if _, err = con.Write(response); err != nil {
			log.Printf("failed to respond to client: %v\n", err)
		}
	}
}

func handleRequest(request []byte) (resp []byte, err error) {
	var params model.BasicParams
	err = json.Unmarshal(request, &params)
	if err != nil {
		return nil, err
	}

	fmt.Println("Handle Request: " + params.RequestType)
	switch params.RequestType {
	case requestType.Login:
		var loginParams model.LogInParams
		if err = json.Unmarshal(request, &loginParams); err != nil {
			//w.WriteHeader(http.StatusBadRequest)
			return createLoginResponse(http.StatusBadRequest, "", time.Now()), err

		}
		expectedPassword, ok := users[loginParams.Password]
		if !ok || expectedPassword != loginParams.Password {
			// w.WriteHeader(http.StatusUnauthorized)
			return createLoginResponse(http.StatusUnauthorized, "", time.Now()), err

		}

		// Create a new random session token
		sessionToken := uuid.NewV4().String()
		// Set the token in the cache, along with the user whom it represents
		// The token has an expiry time of 120 seconds
		_, err = cache.Do("SETEX", sessionToken, "3000", loginParams.Username)
		if err != nil {
			// If there is an error in setting the cache, return an internal server error
			// w.WriteHeader(http.StatusInternalServerError)
			return createLoginResponse(http.StatusInternalServerError, "", time.Now()), err
		}
		return createLoginResponse(http.StatusOK, sessionToken, time.Now().Add(3000*time.Second)), nil
	default:
		err = errors.New("invalid command")
		return []byte{}, err
	}

}

func createLoginResponse(statusCode int, sessionToken string, expireTime time.Time) []byte {
	return []byte(
		fmt.Sprintf(`{"code":%d,"data":{"id":"%s","created_at":"%s", "sessionToken": "%s", "Expires": "%s"}}`+"\n",
			statusCode,
			uuid.NewV4().String(),
			time.Now().UTC().Format(time.RFC3339),
			sessionToken,
			expireTime,
		),
	)
}
