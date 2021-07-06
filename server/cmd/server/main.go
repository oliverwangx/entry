package main

import (
	"bufio"
	"bytes"
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

var usersNickName = map[string]string{
	"user1": "oliver",
	"user2": "Chole",
	"wxy":   "Nancy",
}

var usersAvatarPath = map[string]string{
	"user1": "image/person3.png",
	"user2": "image/person2.png",
	"wxy":   "image/person1.png",
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
				log.Println("handle request error: " + err.Error())
			}

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
		response := &model.LoginResponse{
			Code: http.StatusOK,
			Data: model.LoginData{
				ID:         uuid.NewV4().String(),
				CreatedAt:  time.Now().UTC().Format(time.RFC3339),
				NickName:   "",
				AvatarPath: "",
			},
			SessionToken: "",
			ExpireTime:   time.Now(),
		}
		if err = json.Unmarshal(request, &loginParams); err != nil {
			//w.WriteHeader(http.StatusBadRequest)
			response.Code = http.StatusBadRequest
			return createLoginResponse(response), err

		}
		expectedPassword, ok := users[loginParams.Username]

		if !ok || expectedPassword != loginParams.Password {
			// w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("expectedPassword", expectedPassword, "loginParams Password", loginParams.Password)
			response.Code = http.StatusUnauthorized
			return createLoginResponse(response), err

		}
		// Create a new random session token
		sessionToken := uuid.NewV4().String()
		// Set the token in the cache, along with the user whom it represents
		// The token has an expiry time of 120 seconds
		_, err = cache.Do("SETEX", sessionToken, "30000", loginParams.Username)
		if err != nil {
			// If there is an error in setting the cache, return an internal server error
			// w.WriteHeader(http.StatusInternalServerError)
			response.Code = http.StatusInternalServerError
			return createLoginResponse(response), err
		}
		response.Data.AvatarPath = usersAvatarPath[loginParams.Username]
		response.Data.NickName = usersNickName[loginParams.Username]
		response.Data.SessionToken = sessionToken
		response.SessionToken = sessionToken
		response.ExpireTime = time.Now().UTC().Add(30000 * time.Minute)
		return createLoginResponse(response), nil

	case requestType.UpdateNickname:
		var nickNameParams model.NickNameParams
		if err = json.Unmarshal(request, &nickNameParams); err != nil {
			return createNickNameResponse(http.StatusBadRequest, ""), err
		}
		sessionToken := nickNameParams.SessionToken
		if sessionToken == "" {
			return createNickNameResponse(http.StatusNotFound, ""), http.ErrNoCookie
		}
		userName, err := cache.Do("GET", sessionToken)

		if err != nil {
			// If there is an error fetching from cache, return an internal server error status
			return createNickNameResponse(http.StatusInternalServerError, ""), err
		}
		if userName == nil {
			// If the session token is not present in cache, return an unauthorized error
			return createNickNameResponse(http.StatusBadGateway, ""), err
		}
		// var nickName string
		//a := fmt.Sprintf("%s", userName)
		//fmt.Println("user name", a)
		//if nickName, ok := usersNickName[fmt.Sprintf("%s", userName)]; ok {
		//	return createNickNameResponse(http.StatusOK, nickName), nil
		//}
		//return createNickNameResponse(http.StatusForbidden, ""), nil
		usersNickName[fmt.Sprintf("%s", userName)] = nickNameParams.NickName
		return createNickNameResponse(http.StatusOK, nickNameParams.NickName), nil

	case requestType.UpdateAvatar:
		var AvatarParams model.AvatarParams
		if err = json.Unmarshal(request, &AvatarParams); err != nil {
			return createAvatarResponse(http.StatusBadRequest, ""), err
		}
		sessionToken := AvatarParams.SessionToken
		if sessionToken == "" {
			return createAvatarResponse(http.StatusNotFound, ""), http.ErrNoCookie
		}
		userName, err := cache.Do("GET", sessionToken)

		if err != nil {
			// If there is an error fetching from cache, return an internal server error status
			return createAvatarResponse(http.StatusInternalServerError, ""), err
		}
		if userName == nil {
			// If the session token is not present in cache, return an unauthorized error
			return createAvatarResponse(http.StatusBadGateway, ""), err
		}

		if AvatarPath, ok := usersAvatarPath[fmt.Sprintf("%s", userName)]; ok {
			return createAvatarResponse(http.StatusOK, AvatarPath), nil
		}
		return createAvatarResponse(http.StatusForbidden, ""), nil

	default:
		err = errors.New("invalid command")
		return []byte{}, err
	}

}

// NewResponse creates a network request from a copy of `outgoing` struct.
func NewResponse(outgoing interface{}) (*bytes.Buffer, error) {
	resp := bytes.NewBuffer(nil)
	if err := json.NewEncoder(resp).Encode(outgoing); err != nil {
		return nil, err
	}

	return resp, nil
}

func createLoginResponse(response *model.LoginResponse) []byte {
	resp, err := NewResponse(*response)
	if err != nil {
		fmt.Println("error for the response")
		return []byte{}
	}
	fmt.Println(response)
	clientReader := bufio.NewReader(resp)
	newResp, _ := clientReader.ReadString('\n')
	// fmt.Sprintf("%v", response)
	return []byte(strings.TrimSpace(newResp) + "\n")
}

func createNickNameResponse(statusCode int, nickName string) []byte {
	response := model.NickNameResponse{
		Code: statusCode,
		Data: model.NickNameData{
			NickName: nickName,
		},
	}
	resp, err := NewResponse(response)
	if err != nil {
		fmt.Println("error for the response")
		return []byte{}
	}
	fmt.Println(response)
	clientReader := bufio.NewReader(resp)
	newResp, _ := clientReader.ReadString('\n')
	// fmt.Sprintf("%v", response)
	return []byte(strings.TrimSpace(newResp) + "\n")
}

func createAvatarResponse(statusCode int, AvatarPath string) []byte {
	response := model.AvatarResponse{
		Code: statusCode,
		Data: model.AvatarData{
			AvatarPath: AvatarPath,
		},
	}
	resp, err := NewResponse(response)
	if err != nil {
		fmt.Println("error for the response")
		return []byte{}
	}
	fmt.Println(response)
	clientReader := bufio.NewReader(resp)
	newResp, _ := clientReader.ReadString('\n')
	// fmt.Sprintf("%v", response)
	return []byte(strings.TrimSpace(newResp) + "\n")
}
