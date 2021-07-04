package http

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"shopee-backend-entry-task/client/internal/pkg/storage"
	"shopee-backend-entry-task/client/internal/user"
)

type Router struct {
	handler *httprouter.Router
}

func NewRouter() *Router {
	rtr := httprouter.New()
	rtr.RedirectTrailingSlash = false
	rtr.RedirectFixedPath = false

	return &Router{
		handler: rtr,
	}
}

func (r *Router) RegisterUser(str storage.Storage) {
	r.handler.HandlerFunc(http.MethodPost, "/api/v1/users", user.NewCreate(str).Handle)
}
