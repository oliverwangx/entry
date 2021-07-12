package main

import (
	"net"
	"oliver/entry/client/internal/pkg/http"
	"oliver/entry/client/internal/pkg/storage"
	"oliver/entry/client/internal/pkg/tcp"
	"oliver/entry/config"
	"oliver/entry/utils/logger"
	"oliver/entry/utils/newPool"
)

// var serverConfig map[string]string

func main() {
	// Open the sockets TCP connection pool
	serverConfig, err := config.GetConfig()
	if err != nil {
		logger.Error.Println("Parse Configuration", err)
		return
	}

	factory := func() (net.Conn, error) {
		return tcp.NewConnection(serverConfig[config.TcpHost] + serverConfig[config.TcpPort])
	}
	connectionPool, err := newPool.NewGenericPool(5, 30, 10, factory)
	if err != nil {
		logger.Error.Println("TCP Client Connection Pool Error:", err)
	}
	// close the sockets TCP client
	defer func(pool newPool.Pool) {
		if err := pool.Shutdown(); err != nil {
			logger.Error.Println("Connection Pool shutdown fails")
		}
		logger.Error.Println("Connection Pool shutdown successfully, HTTP server + TCP client shut down")

	}(connectionPool)

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
	//srv.SetKeepAlivesEnabled(false)

	logger.Error.Println(srv.ListenAndServe())
	return
}
