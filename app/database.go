package app

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbConn *pgxpool.Pool

// NewDatabasePool creates a new PostgreSQL connection pool.
// It reads the database configuration from the Config() and establishes a connection.
// It also verifies the connection by executing a simple query.
func NewDatabasePool(ctx context.Context) (db *pgxpool.Pool, err error) {
	c := Config().DB
	db, err = pgxpool.New(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.User, c.Password, c.Host, c.Port, c.Name))
	if err != nil {
		err = fmt.Errorf("create db connection pool: %w", err)
		return
	}

	_, err = db.Exec(ctx, "SELECT 1")
	if err != nil {
		err = fmt.Errorf("db.exec: %w", err)
	}

	dbConn = db
	return
}

// DB returns the global PostgreSQL connection pool.
func DB() *pgxpool.Pool {
	return dbConn
}

// CloseDatabase closes the global PostgreSQL connection pool if it is not nil.
func CloseDatabase() {
	if dbConn != nil {
		dbConn.Close()
	}
}

//go:embed db_migrations/*.sql
var mgFiles embed.FS

const (
	mgTable = "db_migrations"
	mgDir   = "db_migrations"
)

type migration struct {
	Version    int64
	Name       string
	SQL        string
	MigratedAt *time.Time
}

// MigrateDB runs SQL scripts from the './migrations' directory that haven't been committed yet.
// It reads migration files, compares them with the migrations already applied to the database,
// and executes the new migrations in a transaction.
func MigrateDB(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("[ERROR] [MIGRATE]: %v", err)
		}
	}()

	allMg := make([]migration, 0)
	completedMg := make([]migration, 0)
	newMg := make([]migration, 0)

	dir, err := mgFiles.ReadDir(mgDir)
	if err != nil {
		return fmt.Errorf("read migrations: %v", err)
	}

	for _, file := range dir {
		mgData, err := mgFiles.ReadFile(path.Join(mgDir, file.Name()))
		if err != nil {
			return fmt.Errorf("read '%s': %v", file.Name(), err)
		}

		name := strings.TrimSuffix(file.Name(), ".sql")
		idx := strings.Index(name, "_")
		if idx < 0 {
			return fmt.Errorf("%s: no filename separator '_' found", name)
		}

		version, err := strconv.ParseInt(name[:idx], 10, 64)
		if err != nil {
			return fmt.Errorf("parse version from migration file: %s: %w", name, err)
		}

		name = name[idx+1:]

		for _, m := range allMg {
			if m.Version == version {
				return fmt.Errorf("multiple migrations of version %d found", version)
			}
		}

		allMg = append(allMg, migration{
			Version:    version,
			Name:       name,
			SQL:        string(mgData),
			MigratedAt: nil,
		})
	}

	if len(allMg) == 0 {
		fmt.Println("no migrations found in " + mgDir)
		return
	}

selectAll:
	err = pgxscan.Select(ctx, dbConn, &completedMg, `SELECT * FROM `+mgTable)
	if err != nil {
		// on clean db run migrations table not exists yet
		// check this by code returned and table name
		// if so, create table and retry
		pgErr := &pgconn.PgError{}
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" && strings.Contains(err.Error(), mgTable) {
			_, err = dbConn.Exec(ctx, `
       CREATE TABLE `+mgTable+` (
          version    BIGINT NOT NULL PRIMARY KEY,
          name       TEXT NOT NULL,
          migrated_at TIMESTAMPTZ NOT NULL
       );`)
			if err != nil {
				return fmt.Errorf("init migrations table: %v", err)
			}
			goto selectAll
		}
		return err
	}
	if err != nil {
		return fmt.Errorf("load completed migrations: %v", err)
	}

iterate:
	for _, m := range allMg {
		for _, c := range completedMg {
			if m.Version == c.Version {
				continue iterate
			}
		}

		newMg = append(newMg, m)
	}

	if len(newMg) == 0 {
		fmt.Println("[MIGRATION] ✅ Nothing to migrate")
		return
	}

	sort.Slice(newMg, func(i, j int) bool {
		return newMg[i].Version < newMg[j].Version
	})

	for _, m := range newMg {
		err = RunInTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			now := time.Now()
			_, err = tx.Exec(ctx, m.SQL)
			if err != nil {
				return fmt.Errorf("❌ %d %s: %v", m.Version, m.Name, err)
			}

			m.MigratedAt = &now
			_, err = tx.Exec(ctx, "INSERT INTO "+mgTable+" (version, name, migrated_at) VALUES ($1, $2, $3)", m.Version, m.Name, m.MigratedAt)
			if err != nil {
				return fmt.Errorf("❌ save migration %d %s: %v", m.Version, m.Name, err)
			}

			fmt.Printf("[MIGRATION] %d\n✅ %s\n%s\n", m.Version, m.Name, time.Since(now).String())
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// RunInTx executes a function within a PostgreSQL transaction.
// It begins a transaction, executes the provided function, and commits or rolls back the transaction based on the function's result.
func RunInTx(ctx context.Context, f func(ctx context.Context, tx pgx.Tx) error) (err error) {
	tx, err := dbConn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %v", err)
	}

	err = f(ctx, tx)
	if err != nil {
		err2 := tx.Rollback(ctx)
		if err2 != nil {
			err = fmt.Errorf("%v (rollback transaction: %v)", err, err2)
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		err = fmt.Errorf("commit transaction: %v", err)
	}

	return
}
