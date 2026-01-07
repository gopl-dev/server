package cli_test

import (
	"testing"

	"github.com/gopl-dev/server/cli"
)

func TestCommandsRun(t *testing.T) {
	app := cli.NewApp("TestApp", "TEST")

	err := app.Register(
		cli.NewSampleCommandWithSignatureCmd(),
		cli.NewSampleCommandWithNamedParamsCmd(),
		cli.NewSampleCommandWithFlagsCmd(),
	)
	if err != nil {
		t.Fatalf("failed to register commands: %v", err)
	}

	tests := map[string][]struct { // command: []cases
		name      string
		args      []string
		expectErr string
	}{
		"how_to_cli": {
			{"how_to_cli", []string{"Alice", "Excited", "25"}, ""},
			{"how_to_cli_defaults_1", []string{"Alice", "Excited"}, ""},
			{"how_to_cli_defaults_2", []string{"Alice"}, ""},
			{"how_to_cli_required_err", []string{}, "argument 'name' is required"},
		},
		"how_to_params": {
			{"how_to_params", []string{"Bob", "-m=Happy", "-a=30"}, ""},
			{"how_to_params_defaults_1", []string{"Bob", "-m=Happy"}, ""},
			{"how_to_params_defaults_2", []string{"Bob", "-m=Happy"}, ""},
			{"how_to_params_order", []string{"-m=Happy", "-a=30", "Bob"}, ""},
			{"how_to_params_required_err", []string{"-m=Happy"}, "argument 'name' is required"},
		},
		"how_to_flags": {
			{"how_to_flags", []string{"STAGING", "-v", "-y"}, ""},
			{"how_to_flags_defaults", []string{"-y"}, ""},
		},
	}

	for commandName, cases := range tests {
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()

				err := app.Run(commandName, tt.args...)
				if err != nil {
					if tt.expectErr == "" {
						t.Errorf("command %s failed: %v", tt.name, err)
					} else if err.Error() != tt.expectErr {
						t.Errorf("expected error '%s', got '%s'", tt.expectErr, err.Error())
					}
				} else if tt.expectErr != "" {
					t.Errorf("expected error '%s', got nil", tt.expectErr)
				}
			})
		}
	}
}
