package server

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
)

func NewServer(addr string, router *ginext.Engine) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}
