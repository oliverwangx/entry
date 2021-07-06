package main

import (
	"log"
	"net"
	"shopee-backend-entry-task/client/internal/pkg/http"
	"shopee-backend-entry-task/client/internal/pkg/storage"
	"shopee-backend-entry-task/client/internal/pkg/tcp"
	"shopee-backend-entry-task/client/internal/utils/pool"
)

func main() {
	// Open the sockets TCP connection pool
	factory := func() (net.Conn, error) { return tcp.NewConnection("0.0.0.0:8989") }
	connectionPool, err := pool.NewChannelPool(5, 30, factory)
	if err != nil {
		log.Fatalln("TCP Client Connection Error:", err)
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
	srv := http.NewServer("127.0.0.1:8888", *rtr)
	log.Fatalln(srv.ListenAndServe())
}
