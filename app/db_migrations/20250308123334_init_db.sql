CREATE TABLE topics
(
    id         BIGSERIAL PRIMARY KEY NOT NULL,
    name       TEXT UNIQUE           NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE entities
(
    id         BIGSERIAL PRIMARY KEY NOT NULL,
    path       TEXT UNIQUE           NOT NULL,
    title      TEXT                  NOT NULL,
    type       TEXT                  NOT NULL,
    data       JSONB                 NOT NULL,
    created_at TIMESTAMPTZ           NOT NULL,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE TABLE entity_topics
(
    entity_id BIGINT NOT NULL REFERENCES entities (id),
    topic_id  BIGINT NOT NULL REFERENCES topics (id)
);

CREATE TABLE content_import_logs
(
    id         BIGSERIAL PRIMARY KEY NOT NULL,
    status     INT                   NOT NULL,
    log        TEXT                  NOT NULL,
    created_at TIMESTAMPTZ           NOT NULL
);

