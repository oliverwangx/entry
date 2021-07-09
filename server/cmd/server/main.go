package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/satori/go.uuid"
	"io"
	"log"
	"net"
	"net/http"
	"shopee-backend-entry-task/config"
	"shopee-backend-entry-task/model"
	requestType2 "shopee-backend-entry-task/model/requestType"
	Memory2 "shopee-backend-entry-task/server/internal/Memory"
	logger2 "shopee-backend-entry-task/utils/logger"
	"strings"
	"time"
)

var cache redis.Conn
var DataStoreClient Memory2.DataStore
var ctx context.Context

func main() {
	serverConfig, err := config.GetConfig()
	DataStoreClient.Init(serverConfig)
	ctx = context.Background()

	listener, err := net.Listen("tcp", serverConfig[config.TcpHost]+serverConfig[config.TcpPort])
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()

	for {
		con, err := listener.Accept()
		if err != nil {
			log.Println(err)
			logger2.Error.Println("TCP server connection fails", con)
			continue
		}

		go receiveClientRequest(con)
	}
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
				logger2.Error.Println("handle request error: " + err.Error())
			}

			if clientRequest == ":QUIT" {
				logger2.Info.Println("client requested server to close the connection so closing")
				return
			}
		case io.EOF:
			logger2.Error.Println("client closed the connection by terminating the process")
			return
		default:
			logger2.Fatal.Println("error: %v\n", err)
			return
		}

		if _, err = con.Write(response); err != nil {
			logger2.Error.Println("failed to respond to client: %v\n", err)
		}
	}
}

func handleRequest(request []byte) (resp []byte, err error) {
	//serverConfig, err := config.GetConfig()
	if err != nil {
		logger2.Error.Println("Parse Configuration", err)
		return
	}
	var params model.BasicParams
	err = json.Unmarshal(request, &params)
	if err != nil {
		return nil, err
	}

	logger2.Info.Println("Handle Request:" + params.RequestType)
	switch params.RequestType {
	case requestType2.Login:
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
			logger2.Error.Println("TCP server Unmarshal problem", err)
			return createLoginResponse(response), err

		}
		// user := &model.User{}
		user, err := DataStoreClient.GetUserByUsername(ctx, loginParams.Username)
		if err != nil {
			response.Code = http.StatusHTTPVersionNotSupported
			logger2.Error.Println("Joint DataBase Query Error", err)
			return createLoginResponse(response), err
		}
		if user == nil {
			response.Code = http.StatusExpectationFailed
			logger2.Error.Println(":User Receive nil Error", err)
			return createLoginResponse(response), err
		}
		expectedPassword := user.Password
		hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(loginParams.Password)))
		if expectedPassword != hashedPassword {
			// w.WriteHeader(http.StatusUnauthorized)
			response.Code = http.StatusUnauthorized
			return createLoginResponse(response), err

		}
		// Create a new random session token
		sessionToken := uuid.NewV4().String()
		// Set the token in the cache, along with the user whom it represents
		// The token has an expiry time of 3000 seconds
		err = DataStoreClient.Cache.SetUserSession(ctx, loginParams.Username, sessionToken)

		if err != nil {
			// If there is an error in setting the cache, return an internal server error
			// w.WriteHeader(http.StatusInternalServerError)
			response.Code = http.StatusInternalServerError
			return createLoginResponse(response), err
		}
		response.Data.AvatarPath = user.Avatar
		response.Data.NickName = user.Nickname
		response.Data.SessionToken = sessionToken
		response.SessionToken = sessionToken
		//expireTime, err := strconv.Atoi(serverConfig[config.SESSIONTIME])
		//if err != nil{
		//	log.Fatalln("configure format error")
		//}
		response.ExpireTime = time.Now().UTC().Add(30000 * time.Minute)
		return createLoginResponse(response), nil

	case requestType2.UpdateNickname:
		var nickNameParams model.NickNameParams
		if err = json.Unmarshal(request, &nickNameParams); err != nil {
			return createNickNameResponse(http.StatusBadRequest, ""), err
		}
		sessionToken := nickNameParams.SessionToken
		if sessionToken == "" {
			return createNickNameResponse(http.StatusNotFound, ""), http.ErrNoCookie
		}
		userName, err := DataStoreClient.Cache.GetUserSession(ctx, sessionToken)

		if err != nil {
			// If there is an error fetching from cache, return an internal server error status
			return createNickNameResponse(http.StatusInternalServerError, ""), err
		}
		if userName == "" {
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
		err = DataStoreClient.UpdateUserNickname(ctx, userName, nickNameParams.NickName)
		if err != nil {
			return createNickNameResponse(http.StatusExpectationFailed, ""), err
		}
		// usersNickName[fmt.Sprintf("%s", userName)] = nickNameParams.NickName
		return createNickNameResponse(http.StatusOK, nickNameParams.NickName), nil

	case requestType2.UpdateAvatar:
		var AvatarParams model.AvatarParams
		if err = json.Unmarshal(request, &AvatarParams); err != nil {
			return createAvatarResponse(http.StatusBadRequest, ""), err
		}
		sessionToken := AvatarParams.SessionToken
		if sessionToken == "" {
			return createAvatarResponse(http.StatusNotFound, ""), http.ErrNoCookie
		}
		userName, err := DataStoreClient.Cache.GetUserSession(ctx, sessionToken)

		if err != nil {
			// If there is an error fetching from cache, return an internal server error status
			return createNickNameResponse(http.StatusInternalServerError, ""), err
		}

		if userName == "" {
			// If the session token is not present in cache, return an unauthorized error
			return createAvatarResponse(http.StatusBadGateway, ""), err
		}
		//newPath := fmt.Sprintf("image/test2%d.txt", rand.Int())
		//err = os.Rename(AvatarParams.AvatarPath, newPath)
		//if err != nil {
		//	log.Fatal(err)
		//}
		err = DataStoreClient.UpdateUserAvatar(ctx, userName, AvatarParams.AvatarPath)
		if err != nil {
			return createAvatarResponse(http.StatusGatewayTimeout, ""), err

		}
		//usersAvatarPath[fmt.Sprintf("%s", userName)] = AvatarParams.AvatarPath
		return createAvatarResponse(http.StatusOK, AvatarParams.AvatarPath), nil

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
