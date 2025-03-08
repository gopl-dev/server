package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func main() {
	args := make([]string, 0)
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	err := newMigration(args)
	if err != nil {
		println(err.Error())
		return
	}
}

func newMigration(args []string) (err error) {
	name := "new_migration"
	if len(args) > 0 {
		name = strings.ToLower(strings.Join(args, "_"))
	}

	name = time.Now().UTC().Format("20060102150405") + "_" + name + ".sql"
	mg := resolveSampleMg(name)

	err = os.WriteFile("app/db_migrations/"+name, []byte(mg+"\n"), os.ModePerm)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/C start app/db_migrations/"+name)
		err = cmd.Run()
		if err != nil {
			return
		}
	}

	fmt.Println("app/db_migrations/" + name)
	return nil
}

const sampleCreateTableMg = `-- CREATE TABLE my_table
-- (
--     id            uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v1mc(),
--     name          TEXT NOT NULL,
--     workspace_id  BIGINT REFERENCES workspaces (id),
--     user_id       BIGINT REFERENCES users (id),
--     project_id      BIGINT REFERENCES projects (id),
--     
--     created_at    timestamptz NOT NULL,
--     updated_at    timestamptz,
--     deleted_at    timestamptz
-- );

-- CREATE INDEX  my_table_workspace_id_idx ON my_table (workspace_id);
-- CREATE INDEX  my_table_user_id_idx ON my_table (user_id);
-- CREATE INDEX  my_table_project_id_idx ON my_table (project_id);
`

const sampleAddColMg = `-- ALTER TABLE ? ADD COLUMN ? TYPE;`
const sampleDropColMg = `-- ALTER TABLE ? DROP COLUMN ?;`
const sampleIndexMg = `-- CREATE INDEX  ?_idx ON ? (?);`
const sampleRenameMg = `-- ALTER TABLE ? RENAME COLUMN ? TO ?;
-- ALTER TABLE ? RENAME TO ?;`

func resolveSampleMg(name string) string {
	if strings.Contains(name, "create_table") {
		return sampleCreateTableMg
	}

	if strings.Contains(name, "rename") {
		return sampleRenameMg
	}

	if strings.Contains(name, "drop") {
		return sampleDropColMg
	}

	if strings.Contains(name, "add") {
		return sampleAddColMg
	}

	if strings.Contains(name, "index") {
		return sampleIndexMg
	}

	all := []string{
		sampleCreateTableMg,
		sampleAddColMg,
		sampleDropColMg,
		sampleRenameMg,
	}

	return strings.Join(all, "\n\n")
}
