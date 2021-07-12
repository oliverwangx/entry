package tcp

import (
	"net"
	"oliver/entry/utils/logger"
	"time"
)

func NewConnection(address string) (net.Conn, error) {
	Conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		logger.Error.Println("Dial fails")
	}
	return Conn, err
}
