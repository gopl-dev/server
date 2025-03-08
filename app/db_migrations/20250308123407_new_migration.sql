-- CREATE TABLE my_table
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


-- ALTER TABLE ? ADD COLUMN ? TYPE;

-- ALTER TABLE ? DROP COLUMN ?;

-- ALTER TABLE ? RENAME COLUMN ? TO ?;
-- ALTER TABLE ? RENAME TO ?;
