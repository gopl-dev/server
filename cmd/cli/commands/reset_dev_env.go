package commands

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/cli"
	"github.com/gopl-dev/server/trace"
	"github.com/jackc/pgx/v5"
	aur "github.com/logrusorgru/aurora"
	"golang.org/x/crypto/bcrypt"
)

var (
	errInvalidTestDBName = errors.New("invalid test database name")
	errInvalidDBName     = errors.New("invalid database name")
	errInvalidEnv        = errors.New("this command cannot be run in PRODUCTION environment")
)

// NewResetDevEnvCmd ...
func NewResetDevEnvCmd() cli.Command {
	return cli.Command{
		Name:  "reset_dev_env",
		Alias: "rde",
		Help: []string{
			"Reset development environment",
			"Drop & create DB, run migrations, create default user, drop & create test DB (if found)",
			"To make it work, your DB must be named with '_local_dev' suffix (ex: proj_local_dev) and your test DB must be named as '{devdb}_test' (ex: proj_local_dev_test)",
			"-u: Username for new user",
			"-e: Email for new user",
			"-p: Password for new user",
			"-ns: Do not seed data",
		},
		Handler: &resetDevEnvCmd{},
	}
}

type resetDevEnvCmd struct {
	NoSeed   bool    `arg:"-ns"`
	Username *string `arg:"-u" default:"admin"`
	Email    *string `arg:"-e" default:"admin"`
	Password *string `arg:"-p" default:"admin"`
}

func (cmd *resetDevEnvCmd) Handle(ctx context.Context) error {
	if app.Config().IsProductionEnv() {
		return errInvalidEnv
	}

	const (
		devDbNameSuffix  = "_local_dev"
		testDbNameSuffix = "_test"
	)

	conf := app.Config().DB
	if !strings.HasSuffix(conf.Name, devDbNameSuffix) {
		return fmt.Errorf("%w: expected suffix '_dev', got %s", errInvalidDBName, devDbNameSuffix)
	}

	// Load test config now, to abort early
	// I assume that test DB might not created yet,
	// also one test DB is used for all test namespaces (api, services, workers)
	var (
		err        error
		testConf   *app.ConfigT
		testDbName = conf.Name + testDbNameSuffix
	)
	for _, name := range []string{"api", "service", "worker"} {
		confPath := fmt.Sprintf("./test/%s_test/.config.yaml", name)
		testConf, err = app.ConfigFromFile(confPath)
		if errors.Is(err, &os.PathError{}) {
			continue
		}
		if err != nil {
			return fmt.Errorf("load test config (%s): %w", confPath, err)
		}
		if testConf.DB.Name == "" {
			continue
		}
		fmt.Println("TEST CONFIG LOADED FROM: " + confPath)
		fmt.Println("TEST DB: " + testConf.DB.Name)

		if testConf.DB.Name != testDbName {
			return fmt.Errorf("%w: expected %s, got %s", errInvalidTestDBName, testDbName, testConf.DB.Name)
		}

		break
	}

	// Close existing connection to allow dropping the DB
	CloseDB()

	// Connect to postgres DB to perform admin tasks
	hostPort := net.JoinHostPort(conf.Host, conf.Port)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/postgres",
		conf.User, conf.Password, hostPort,
	)

	pg, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect to postgres: %w", err)
	}
	defer func() { _ = pg.Close(ctx) }()

	// Helper to drop/create DB
	recreateDB := func(name string) error {
		fmt.Printf("Recreating DB %s...\n", name)
		_, err := pg.Exec(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s" WITH (FORCE)`, name))
		if err != nil {
			return fmt.Errorf("drop db %s: %w", name, err)
		}
		_, err = pg.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, name))
		if err != nil {
			return fmt.Errorf("create db %s: %w", name, err)
		}

		return nil
	}

	// Drop & create DB
	err = recreateDB(conf.Name)
	if err != nil {
		return err
	}

	// Migrate DB
	// We need to reconnect to the new Main DB
	newDB, err := app.NewDB(ctx)
	if err != nil {
		return fmt.Errorf("connect to new db: %w", err)
	}
	defer newDB.Close()

	fmt.Println("Migrating DB...")
	err = app.MigrateDB(ctx, newDB)
	if err != nil {
		return fmt.Errorf("migrate db: %w", err)
	}

	// Create user
	r := repo.New(newDB, trace.NewNoOpTracer())
	fmt.Println("Creating default user...")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(*cmd.Password), bcrypt.MinCost)
	if err != nil {
		return err
	}

	u := &ds.User{
		ID:             ds.NewID(),
		Username:       *cmd.Username,
		Email:          *cmd.Email,
		EmailConfirmed: true,
		Password:       string(passwordHash),
		CreatedAt:      time.Now(),
		UpdatedAt:      nil,
		DeletedAt:      nil,
		IsAdmin:        false,
	}
	err = r.CreateUser(ctx, u)
	if err != nil {
		return err
	}

	err = writeAdminsToConfig(u.ID.String())
	if err != nil {
		return fmt.Errorf("update admins in config: %w", err)
	}

	// Drop & create test DB
	if testConf != nil {
		err = recreateDB(testDbName)
		if err != nil {
			return err
		}
	}

	// Seed data
	if !cmd.NoSeed {
		fmt.Println("Seeding data...")
		seedCmd := seedDataCmd{
			Data:  new("all"),
			Count: new(100), //nolint:mnd
		}
		err = seedCmd.Handle(ctx)
		if err != nil {
			return err
		}
	}

	cli.OK("New user %s  created\n\tEmail: %s\n\tPassword: %s", aur.Bold("with admin role"), *cmd.Email, *cmd.Password)
	if !cmd.NoSeed {
		cli.OK("New test user created\n\tEmail: test\n\tPassword: test")
	}

	return nil
}

func writeAdminsToConfig(adminID string) error {
	confPath := "./.config.yaml"
	data, err := os.ReadFile(confPath)
	if err != nil {
		return err
	}

	content := string(data)

	eol := "\n"
	if strings.Contains(content, "\r\n") {
		eol = "\r\n"
	}

	lines := strings.Split(content, eol)

	adminsIdx := -1
	nextRootIdx := len(lines)
	newAdmins := fmt.Sprintf("admins: ['%s']", adminID)

	for i, line := range lines {
		if strings.HasPrefix(line, "admins:") {
			adminsIdx = i
			break
		}
	}

	// find next root key
	if adminsIdx >= 0 {
		for i := adminsIdx + 1; i < len(lines); i++ {
			if len(lines[i]) == 0 {
				continue
			}

			switch lines[i][0] {
			case ' ', '\t', '#':
				continue
			}

			nextRootIdx = i
			break
		}

		// delete everything in between
		lines = append(
			lines[:adminsIdx+1],
			lines[nextRootIdx:]...,
		)

		lines[adminsIdx] = newAdmins
	} else {
		// add new root key at the end
		if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) != "" {
			lines = append(lines, "")
		}
		lines = append(lines, newAdmins)
	}

	out := strings.Join(lines, eol)
	if !strings.HasSuffix(out, eol) {
		out += eol
	}

	// write new data to config file
	// TODO replace with utility, when files driver will be ready
	dir := filepath.Dir(confPath)
	tmp, err := os.CreateTemp(dir, ".config.yaml.tmp")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()

	_, err = tmp.WriteString(out)
	if err != nil {
		_ = tmp.Close()
		return err
	}
	err = tmp.Close()
	if err != nil {
		return err
	}

	st, err := os.Stat(confPath)
	if err == nil {
		_ = os.Chmod(tmp.Name(), st.Mode())
	}

	return os.Rename(tmp.Name(), confPath)
}
