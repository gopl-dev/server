// Package factory ...
package factory

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"dario.cat/mergo"
	"github.com/gopl-dev/server/app/repo"
)

// Factory holds dependencies required by factory methods.
type Factory struct {
	repo *repo.Repo
}

// New is a factory function that creates and returns a new Factory instance.
func New(r *repo.Repo) *Factory {
	return &Factory{repo: r}
}

func merge(dst, src any) {
	err := mergo.Merge(dst, src, mergo.WithOverride, mergo.WithTransformers(timeTransformer{}))
	if err != nil {
		panic(fmt.Sprintf("Unable to merge %T with %T", dst, src))
	}
}

type timeTransformer struct{}

func (t timeTransformer) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeFor[time.Time]() {
		return func(dst, src reflect.Value) error {
			if src.CanInterface() {
				timeVal, ok := src.Interface().(time.Time)
				if ok && !timeVal.IsZero() {
					if dst.CanSet() {
						dst.Set(src)
					}
				}
			}
			return nil
		}
	}
	return nil
}

// X is a function that repeatedly executes a data creation function ('fn')
// a specified number of times and returns a slice of pointers to the created objects.
//
// T is the type of the struct being created (e.g., ds.User).
// fn is the function that creates a single instance of T (e.g., CreateUser).
// override allows passing custom field values to override defaults in the created instances.
func X[T any](t *testing.T, times int, fn func(t *testing.T, m ...T) *T, override ...T) []*T {
	t.Helper()

	data := make([]*T, times)
	for i := range data {
		data[i] = fn(t, override...)
	}

	return data
}

// Two is a convenience function to create exactly two instances of a data structure T.
// It is a wrapper around X with times=2.
func Two[T any](t *testing.T, fn func(t *testing.T, m ...T) *T, override ...T) []*T {
	t.Helper()

	return X(t, 2, fn, override...) //nolint:mnd
}

// Five is a convenience function to create exactly five instances of a data structure T.
// It is a wrapper around X with times=5.
func Five[T any](t *testing.T, fn func(t *testing.T, m ...T) *T, override ...T) []*T {
	t.Helper()

	return X(t, 5, fn, override...) //nolint:mnd
}

// Ten is a convenience function to create exactly ten instances of a data structure T.
// It is a wrapper around X with times=10.
func Ten[T any](t *testing.T, fn func(t *testing.T, m ...T) *T, override ...T) []*T {
	t.Helper()

	return X(t, 10, fn, override...) //nolint:mnd
}

// checkErr is a test helper function that fails the test immediately if the provided error is not nil.
func checkErr(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}
