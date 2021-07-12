package http

import (
	"net/http"
	"time"
)

func NewServer(adr string, rtr Router) http.Server {
	return http.Server{
		Addr:    adr,
		Handler: rtr.handler,
		ReadTimeout:  300 * time.Microsecond,
		WriteTimeout: 300 * time.Microsecond,
	}
}
