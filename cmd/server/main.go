package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gopl-dev/server/config"
	"github.com/gopl-dev/server/server"
)

func main() {
	conf := config.Get()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGTERM,
		syscall.SIGHUP,  // kill -SIGHUP
		syscall.SIGINT,  // kill -SIGINT or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT
	)

	srv := server.New()

	go func() {
		<-quit
		if err := srv.Close(); err != nil {
			log.Fatal(err.Error())
		}
	}()

	log.Println(conf.App.Name + " (" + conf.App.Version + ") serving at " + conf.Server.Host + ":" + conf.Server.Port)
	err := srv.ListenAndServe()
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err.Error())
	}

	log.Println(conf.App.Name + " (" + conf.App.Version + ") server closed")
}
