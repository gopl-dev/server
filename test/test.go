package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/email"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logrusorgru/aurora"
)

type Data map[string]any
type notNull bool

var NotNull notNull

func CheckErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// StructToQueryString converts flat struct to query string
func StructToQueryString(in interface{}) (out string, err error) {
	jsonBytes, err := json.Marshal(in)
	if err != nil {
		return
	}

	data := map[string]interface{}{}
	err = json.Unmarshal(jsonBytes, &data)
	if err != nil {
		return
	}

	vals := url.Values{}
	for key, v := range data {
		vSlice, ok := v.([]interface{})
		if ok {
			for _, s := range vSlice {
				vals.Add(key, fmt.Sprintf("%v", s))
			}
			continue
		}
		vals.Add(key, fmt.Sprintf("%v", v))
	}

	return vals.Encode(), nil
}

func LoadEmailVars(t *testing.T, to string) map[string]any {
	c, err := email.LoadTestEmail(to)
	if err != nil {
		t.Error(err)
	}

	return c.Variables()
}

func dbIdent(i string) string {
	return pgx.Identifier{i}.Sanitize()
}

func countDatabaseRows(t *testing.T, db *pgxpool.Pool, table string, data Data) int {
	args := make([]interface{}, 0)
	wheres := make([]string, 0)

	argIndex := 1
	for col, val := range data {
		col = dbIdent(col)
		if val == nil {
			wheres = append(wheres, fmt.Sprintf(`"%s" IS NULL`, col))
			continue
		}

		var whereExpr string

		switch val.(type) {
		case bool:
			whereExpr = fmt.Sprintf(`%s IS %v`, col, val)
		case notNull:
			whereExpr = col + " IS NOT NULL"
		default:
			whereExpr = fmt.Sprintf(`%s = $%d`, col, argIndex)
			args = append(args, val)
			argIndex++
		}

		wheres = append(wheres, whereExpr)
	}

	query := fmt.Sprintf("SELECT COUNT(1) AS COUNT FROM %s WHERE %s", dbIdent(table), strings.Join(wheres, " AND "))

	var count int
	err := pgxscan.Get(context.Background(), db, &count, query, args...)
	if err != nil {
		t.Errorf("CountDatabaseRows: %s", err)
	}

	return count
}

func AssertInDB(t *testing.T, db *pgxpool.Pool, table string, data Data) {
	count := countDatabaseRows(t, db, table, data)
	if count == 0 {
		t.Fail()
		println(aurora.Bold(aurora.Red("❌ Table '" + table + "' missing row with data:")).String())
		for k, v := range data {
			println("\t" + k + ": " + fmt.Sprintf("%+v", v))
		}
	}
}

func AssertNotInDB(t *testing.T, db *pgxpool.Pool, table string, data Data) {
	count := countDatabaseRows(t, db, table, data)
	if count != 0 {
		t.Fail()
		println(aurora.Bold(aurora.Red("❌ Table '" + table + "' has row with data:")).String())
		for k, v := range data {
			println("\t" + k + ": " + fmt.Sprintf("%+v", v))
		}
	}
}
