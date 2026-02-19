// Package main ...
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
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/file"
	"github.com/gopl-dev/server/server"
	"github.com/gopl-dev/server/trace"
	"github.com/gopl-dev/server/worker"
)

func main() {
	_ = file.Storage()
	conf := app.Config()
	ctx, cancelCtx := context.WithCancel(context.Background())
	tracer, err := trace.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = worker.Start(ctx)
	if err != nil {
		log.Fatal(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGTERM,
		syscall.SIGHUP,  // kill -SIGHUP
		syscall.SIGINT,  // kill -SIGINT or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT
	)

	db, err := app.NewDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = app.MigrateDB(ctx, db)
	if err != nil {
		log.Fatal(err)
	}

	services := service.New(db, tracer)
	srv := server.New(services, tracer)

	go func() {
		<-quit
		cancelCtx()
		err := srv.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println(conf.App.Name + " (" + conf.App.Version + ") serving at " + conf.Server.Host + ":" + conf.Server.Port)

	if conf.Server.AutocertHosts != "" {
		err = srv.ListenAndServeTLS("", "")
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println(err.Error())
		return
	}

	log.Println(conf.App.Name + " (" + conf.App.Version + ") server closed")
}
