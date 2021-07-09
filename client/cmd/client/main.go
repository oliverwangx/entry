package main

import (
	"log"
	"net"
	"shopee-backend-entry-task/client/internal/pkg/http"
	"shopee-backend-entry-task/client/internal/pkg/storage"
	"shopee-backend-entry-task/client/internal/pkg/tcp"
	"shopee-backend-entry-task/config"
	logger2 "shopee-backend-entry-task/utils/logger"
	pool2 "shopee-backend-entry-task/utils/pool"
)

// var serverConfig map[string]string

func main() {
	// Open the sockets TCP connection pool
	serverConfig, err := config.GetConfig()
	if err != nil {
		logger2.Error.Println("Parse Configuration", err)
		return
	}

	factory := func() (net.Conn, error) {
		return tcp.NewConnection(serverConfig[config.TcpHost] + serverConfig[config.TcpPort])
	}
	connectionPool, err := pool2.NewChannelPool(500, 2000, factory)
	if err != nil {
		logger2.Error.Println("TCP Client Connection Error:", err)
	}
	// close the sockets TCP client
	defer connectionPool.Close()

	// storage binds the TCP connections
	str := storage.New(connectionPool)

	// start new router and register function
	// 1. Register user log in

	rtr := http.NewRouter()
	rtr.RegisterUser(str)
	rtr.RegisterNickName(str)
	rtr.RegisterAvatar(str)

	// Open the http server, listening and serving
	srv := http.NewServer(serverConfig[config.WebHost]+serverConfig[config.WebPort], *rtr)
	log.Fatalln(srv.ListenAndServe())
}
