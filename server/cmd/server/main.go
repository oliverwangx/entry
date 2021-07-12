package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	"io"
	"log"
	"net"
	"oliver/entry/config"
	"oliver/entry/model"
	"oliver/entry/model/requestType"
	"oliver/entry/server/internal/Memory"
	"oliver/entry/server/internal/Service/Auth"
	"oliver/entry/server/internal/Service/Avatar"
	"oliver/entry/server/internal/Service/NickName"
	"oliver/entry/utils/logger"
	"strings"
)

var cache redis.Conn
var DataStoreClient Memory.DataStore
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
			logger.Error.Println("TCP server connection fails", con)
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
				logger.Error.Println("handle request error: " + err.Error())
			}

			if clientRequest == ":QUIT" {
				//logger2.Info.Println("client requested server to close the connection so closing")
				return
			}
		case io.EOF:
			logger.Error.Println("client closed the connection by terminating the process")
			return
		default:
			logger.Fatal.Println("error: %v\n", err)
			return
		}

		if _, err = con.Write(response); err != nil {
			logger.Error.Println("failed to respond to client: %v\n", err)
		}
	}
}

func handleRequest(request []byte) (resp []byte, err error) {
	var params model.BasicParams
	err = json.Unmarshal(request, &params)
	if err != nil {
		return nil, err
	}
	logger.Info.Println("Current Handle Request:" + params.RequestType)
	switch params.RequestType {
	case requestType.Login:
		return Auth.LoginInService(request, DataStoreClient, ctx)

	case requestType.UpdateNickname:
		return NickName.NickNameService(request, DataStoreClient, ctx)

	case requestType.UpdateAvatar:
		return Avatar.AvatarService(request, DataStoreClient, ctx)

	default:
		err = errors.New("invalid command")
		return []byte{}, err
	}

}