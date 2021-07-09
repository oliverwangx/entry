package tcp

import (
	"net"
	"time"
)

func NewConnection(address string) (net.Conn, error) {
	return net.DialTimeout("tcp", address, 100*time.Second)
}
