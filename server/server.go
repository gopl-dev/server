package server

import (
	"net"
	"net/http"

	"github.com/gopl-dev/server/config"
)

func New() *http.Server {
	conf := config.Get()

	return &http.Server{
		Addr:    net.JoinHostPort(conf.Server.Host, conf.Server.Port),
		Handler: Handler(),
	}
}
