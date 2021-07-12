package Auth

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"oliver/entry/model"
	"oliver/entry/server/internal/Memory"
	"oliver/entry/utils/logger"
	"strings"
	"time"
)
func CreateFakeLoginResp()*model.LoginResponse {
	return &model.LoginResponse{
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
}


func LoginInService(request []byte, DataStoreClient Memory.DataStore, ctx context.Context) (resp []byte, err error) {
	var loginParams model.LogInParams
	response := CreateFakeLoginResp()
	if err = json.Unmarshal(request, &loginParams); err != nil {
		response.Code = http.StatusBadRequest
		logger.Error.Println("TCP server Unmarshal problem", err)
		return sendLoginResponse(response), err

	}

	user, err := DataStoreClient.GetUserByUsername(ctx, loginParams.Username)
	if err != nil {
		response.Code = http.StatusHTTPVersionNotSupported
		logger.Error.Println("Joint DataBase Query Error", err)
		return sendLoginResponse(response), err
	}

	if user == nil {
		response.Code = http.StatusExpectationFailed
		logger.Error.Println(":User Receive nil Error", err)
		return sendLoginResponse(response), err
	}

	hashedPassword := fmt.Sprintf("%x", md5.Sum([]byte(loginParams.Password)))
	if user.Password != hashedPassword {
		response.Code = http.StatusUnauthorized
		return sendLoginResponse(response), err

	}

	// Create a new random session token
	sessionToken := uuid.NewV4().String()
	if err = DataStoreClient.Cache.SetUserSession(ctx, loginParams.Username, sessionToken); err != nil {
		response.Code = http.StatusInternalServerError
		return sendLoginResponse(response), err
	}

	response = createRealLoginResponse(user, sessionToken)
	return sendLoginResponse(response), nil
}

func createRealLoginResponse(user *model.User, sessionToken string) *model.LoginResponse{
	return &model.LoginResponse{
		Code: http.StatusOK,
		Data: model.LoginData{
			ID:         uuid.NewV4().String(),
			CreatedAt:  time.Now().UTC().Format(time.RFC3339),
			NickName:  user.Nickname ,
			AvatarPath: user.Avatar,
			SessionToken:sessionToken,
		},
		SessionToken: sessionToken,
		ExpireTime:  time.Now().UTC().Add(30000 * time.Minute),
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

func sendLoginResponse(response *model.LoginResponse) []byte {
	resp, err := NewResponse(*response)
	if err != nil {
		fmt.Println("error for the response")
		return []byte{}
	}
	logger.Info.Println("return tcp response is ", response)
	clientReader := bufio.NewReader(resp)
	newResp, _ := clientReader.ReadString('\n')
	return []byte(strings.TrimSpace(newResp) + "\n")
}