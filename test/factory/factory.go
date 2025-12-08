package factory

import (
	"fmt"
	"testing"

	"dario.cat/mergo"
	"github.com/gopl-dev/server/app/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Factory struct {
	repo *repo.Repo
}

func New(db *pgxpool.Pool) *Factory {
	return &Factory{repo: repo.New(db)}
}

func merge(dst, src any) {
	err := mergo.Merge(dst, src, mergo.WithOverride)
	if err != nil {
		panic(fmt.Sprintf("Unable to merge %T with %T", dst, src))
	}
}

func X[T any](t *testing.T, times int, fn func(t *testing.T, m ...T) *T, override ...T) []*T {
	data := make([]*T, times)
	for i := range data {
		data[i] = fn(t, override...)
	}

	return data
}

func Two[T any](t *testing.T, fn func(t *testing.T, m ...T) *T, override ...T) []*T {
	return X(t, 2, fn, override...)
}

func Five[T any](t *testing.T, fn func(t *testing.T, m ...T) *T, override ...T) []*T {
	return X(t, 5, fn, override...)
}

func Ten[T any](t *testing.T, fn func(t *testing.T, m ...T) *T, override ...T) []*T {
	return X(t, 10, fn, override...)
}

func DerefArray[T any](arr []*T) (r []T) {
	r = make([]T, 0)
	for _, v := range arr {
		if v != nil {
			r = append(r, *v)
		}
	}

	return
}
