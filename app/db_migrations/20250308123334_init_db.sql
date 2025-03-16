CREATE TABLE users
(
    id              BIGSERIAL PRIMARY KEY NOT NULL,
    username        TEXT UNIQUE           NOT NULL,
    email           TEXT UNIQUE           NOT NULL,
    password        TEXT                  NOT NULL,
    email_confirmed BOOL,
    created_at      TIMESTAMPTZ           NOT NULL,
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE email_confirmations
(
    id         BIGSERIAL PRIMARY KEY NOT NULL,
    user_id    BIGINT                NOT NULL REFERENCES users (id),
    code       TEXT UNIQUE           NOT NULL,
    created_at TIMESTAMPTZ           NOT NULL,
    expires_at TIMESTAMPTZ           NOT NULL
);

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


