package validation_test

import (
	"errors"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/gopl-dev/server/app"
)

func checkValidatedInput(t *testing.T, valid bool, err error, argName string, expectedErr string) {
	t.Helper()

	if valid && err != nil {
		t.Error(err)
	}

	if valid {
		return
	}

	if err == nil {
		t.Fatal("expected error")
	}

	var inputError app.InputError
	if !errors.As(err, &inputError) {
		t.Error("[ERROR]: ", err.Error())
		t.Fatalf("error expected to be of type InputError, %T given", err)
	}

	errValue, ok := inputError[argName]
	if !ok {
		t.Fatalf("failed to find key '%s' in error message", argName)
	}

	assert.Equal(t, expectedErr, errValue)
}
