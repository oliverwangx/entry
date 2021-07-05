package http

import (
	"net/http"
)

func NewServer(adr string, rtr Router) http.Server {
	return http.Server{
		Addr:    adr,
		Handler: rtr.handler,
	}
}