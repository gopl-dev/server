// Package commands ...
package commands

import (
	"context"
	"sync"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/trace"
)

var (
	onceServices     sync.Once
	onceDB           sync.Once
	onceRepo         sync.Once
	servicesInstance *service.Service
	dbInstance       *app.DB
	repoInstance     *repo.Repo
)

func db() *app.DB {
	onceDB.Do(func() {
		var err error
		dbInstance, err = app.NewDB(context.Background())
		if err != nil {
			panic(err)
		}
	})

	return dbInstance
}

func services() *service.Service {
	onceServices.Do(func() {
		servicesInstance = service.New(db(), trace.NewNoOpTracer())
	})

	return servicesInstance
}

func repos() *repo.Repo {
	onceRepo.Do(func() {
		repoInstance = repo.New(db(), trace.NewNoOpTracer())
	})

	return repoInstance
}

// CloseDB closes the database connection.
func CloseDB() {
	if dbInstance != nil {
		dbInstance.Close()
	}
}
