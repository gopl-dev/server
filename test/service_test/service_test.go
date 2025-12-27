package service_test

import (
	"os"
	"testing"

	"github.com/gopl-dev/server/test"
)

var tt *test.App

func TestMain(m *testing.M) {
	tt = test.NewApp()

	code := m.Run()

	tt.Shutdown()
	os.Exit(code)
}
