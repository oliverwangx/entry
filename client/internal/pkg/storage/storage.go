package storage

import (
	"bufio"
	"github.com/pkg/errors"
	"io"
	"net"
	"shopee-backend-entry-task/client/internal/utils/pool"
	"shopee-backend-entry-task/logger"
	"strings"
)

type Storage struct {
	connectionPool pool.Pool
}

func New(p pool.Pool) Storage {
	return Storage{
		connectionPool: p,
	}
}

// Store writes data to network connection for handling. It prepares a
// network request from a copy of `outgoing` struct and updates the
// reference of `incoming` struct with response body.
func (s Storage) Store(outgoing interface{}, incoming interface{}) error {
	req, err := NewRequest(outgoing)
	if err != nil {
		return errors.Wrap(err, "create request")
	}
	clientReader := bufio.NewReader(req)
	TCPConn, err := s.connectionPool.Get()
	if err != nil {
		return errors.Wrap(err, "TCP connection get method failed")
	}
	defer func(TCPConn net.Conn) {
		err := TCPConn.Close()
		if err != nil {
			logger.Error.Println("TCP connection return to the pool failed")
		}
	}(TCPConn)

	serverReader := bufio.NewReader(TCPConn)

	for {
		req, err := clientReader.ReadString('\n')
		switch err {
		case nil:
			if _, err = TCPConn.Write([]byte(strings.TrimSpace(req) + "\n")); err != nil {
				return errors.Wrap(err, "send request")
			}
		case io.EOF:
			return errors.Wrap(err, "client closed the connection")
		default:
			return errors.Wrap(err, "client")
		}

		res, err := serverReader.ReadString('\n')
		logger.Info.Println("Receive HTTP Response : ", res)
		switch err {
		case nil:
			return NewResponse(res, incoming)
		case io.EOF:
			return errors.Wrap(err, "server closed the connection")
		default:
			return errors.Wrap(err, "server")
		}
	}
}
