package tcp

import (
	"net"
	logger2 "shopee-backend-entry-task/utils/logger"
	"time"
)

func NewConnection(address string) (net.Conn, error) {
	Conn, err := net.DialTimeout("tcp", address, 100*time.Second)
	if err != nil {
		logger2.Error.Println("Dial fails")
	}
	return Conn, err
}
