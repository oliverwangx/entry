package NickName

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"oliver/entry/model"
	"oliver/entry/server/internal/Memory"
	"strings"
)

func NickNameService(request []byte,  DataStoreClient Memory.DataStore, ctx context.Context) (resp []byte, err error){
	var nickNameParams model.NickNameParams
	if err = json.Unmarshal(request, &nickNameParams); err != nil {
		return sendNickNameResponse(http.StatusBadRequest, ""), err
	}
	sessionToken := nickNameParams.SessionToken
	if sessionToken == "" {
		return sendNickNameResponse(http.StatusNotFound, ""), http.ErrNoCookie
	}
	userName, err := DataStoreClient.Cache.GetUserSession(ctx, sessionToken)

	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		return sendNickNameResponse(http.StatusInternalServerError, ""), err
	}
	if userName == "" {
		// If the session token is not present in cache, return an unauthorized error
		return sendNickNameResponse(http.StatusBadGateway, ""), err
	}

	err = DataStoreClient.UpdateUserNickname(ctx, userName, nickNameParams.NickName)
	if err != nil {
		return sendNickNameResponse(http.StatusExpectationFailed, ""), err
	}
	// usersNickName[fmt.Sprintf("%s", userName)] = nickNameParams.NickName
	return sendNickNameResponse(http.StatusOK, nickNameParams.NickName), nil
}

// NewResponse creates a network request from a copy of `outgoing` struct.
func NewResponse(outgoing interface{}) (*bytes.Buffer, error) {
	resp := bytes.NewBuffer(nil)
	if err := json.NewEncoder(resp).Encode(outgoing); err != nil {
		return nil, err
	}

	return resp, nil
}

func sendNickNameResponse(statusCode int, nickName string) []byte {
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