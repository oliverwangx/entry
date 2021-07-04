package tcp

import (
	"net"
)

func NewConnection(address string) (net.Conn, error) {
	return net.Dial("tcp", address)
}
