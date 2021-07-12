package Avatar

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


func AvatarService(request []byte, DataStoreClient Memory.DataStore, ctx context.Context) (resp []byte, err error) {
	var AvatarParams model.AvatarParams
	if err = json.Unmarshal(request, &AvatarParams); err != nil {
		return sendAvatarResponse(http.StatusBadRequest, ""), err
	}
	sessionToken := AvatarParams.SessionToken
	if sessionToken == "" {
		return sendAvatarResponse(http.StatusNotFound, ""), http.ErrNoCookie
	}
	userName, err := DataStoreClient.Cache.GetUserSession(ctx, sessionToken)

	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		return sendAvatarResponse(http.StatusInternalServerError, ""), err
	}

	if userName == "" {
		// If the session token is not present in cache, return an unauthorized error
		return sendAvatarResponse(http.StatusBadGateway, ""), err
	}
	err = DataStoreClient.UpdateUserAvatar(ctx, userName, AvatarParams.AvatarPath)
	if err != nil {
		return sendAvatarResponse(http.StatusGatewayTimeout, ""), err

	}
	//usersAvatarPath[fmt.Sprintf("%s", userName)] = AvatarParams.AvatarPath
	return sendAvatarResponse(http.StatusOK, AvatarParams.AvatarPath), nil
}

func sendAvatarResponse(statusCode int, AvatarPath string) []byte {
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

// NewResponse creates a network request from a copy of `outgoing` struct.
func NewResponse(outgoing interface{}) (*bytes.Buffer, error) {
	resp := bytes.NewBuffer(nil)
	if err := json.NewEncoder(resp).Encode(outgoing); err != nil {
		return nil, err
	}

	return resp, nil
}