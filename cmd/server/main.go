package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/api"
)

func main() {
	conf := app.Config()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGTERM,
		syscall.SIGHUP,  // kill -SIGHUP
		syscall.SIGINT,  // kill -SIGINT or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT
	)

	ctx := context.Background()
	db, err := app.NewDatabasePool(ctx)
	defer db.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = app.MigrateDB(ctx)
	if err != nil {
		log.Fatal(err)
	}

	srv := api.NewServer()
	go func() {
		<-quit
		if err := srv.Close(); err != nil {
			log.Fatal(err.Error())
		}
	}()

	log.Println(conf.App.Name + " (" + conf.App.Version + ") serving at " + conf.Server.Host + ":" + conf.Server.Port)
	err = srv.ListenAndServe()
	if err != nil && errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err.Error())
	}

	log.Println(conf.App.Name + " (" + conf.App.Version + ") server closed")
}
