package storage

import (
	"bufio"
	"github.com/pkg/errors"
	"io"
	"net"
	"oliver/entry/utils/logger"
	"oliver/entry/utils/newPool"
	"strings"
)

type Storage struct {
	connectionPool newPool.Pool
}

func New(p newPool.Pool) Storage {
	return Storage{
		connectionPool: p,
	}
}

// Store writes data to network connection for handling. It prepares a
// network request from a copy of `outgoing` struct and updates the
// reference of `incoming` struct with response body.

func (s Storage) Store(outgoing interface{}, incoming interface{}) error {


	req, err := NewRequest(outgoing)
	defer func(err error) {
		if err != nil{
			logger.Error.Println(err)
		}
	}(err)

	if err != nil {
		return errors.Wrap(err, "create request")
	}
	clientReader := bufio.NewReader(req)
	TCPConn, err := s.connectionPool.Acquire()
	if err != nil {
		return errors.Wrap(err, "TCP connection get method failed")
	}


	defer func(connectionPool newPool.Pool, conn net.Conn) {
		err := connectionPool.Release(conn)
		if err != nil {
			logger.Error.Println("Release TCP conn to Pool Error")
		}
	}(s.connectionPool, TCPConn)

	serverReader := bufio.NewReader(TCPConn)

	for {
		eachReq, err := clientReader.ReadString('\n')
		switch err {
		case nil:
			//logger2.Info.Println("Successfully received the request:", eachReq)
			if _, err = TCPConn.Write([]byte(strings.TrimSpace(eachReq) + "\n")); err != nil {
				logger.Error.Fatalln("TCP client send the request failed, TCPConn", TCPConn)
				return errors.Wrap(err, "send request failed")
			}
		case io.EOF:
			return errors.Wrap(err, "client closed the connection")
		default:
			return errors.Wrap(err, "client")
		}

		res, err := serverReader.ReadString('\n')
			//logger2.Info.Println("Receive TCP Response : ", res)
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
