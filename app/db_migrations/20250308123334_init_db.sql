CREATE TABLE users
(
    id              UUID PRIMARY KEY NOT NULL,
    username        TEXT UNIQUE      NOT NULL,
    email           TEXT UNIQUE      NOT NULL,
    password        TEXT             NOT NULL,
    email_confirmed BOOL,
    created_at      TIMESTAMPTZ      NOT NULL,
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE email_confirmations
(
    id         UUID PRIMARY KEY NOT NULL,
    user_id    UUID             NOT NULL REFERENCES users (id),
    code       TEXT UNIQUE      NOT NULL,
    created_at TIMESTAMPTZ      NOT NULL,
    expires_at TIMESTAMPTZ      NOT NULL
);

CREATE TABLE user_sessions
(
    id         UUID PRIMARY KEY NOT NULL,
    user_id    UUID             NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ      NOT NULL,
    updated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ      NOT NULL
);


CREATE TABLE password_reset_tokens
(
    id         UUID PRIMARY KEY NOT NULL,
    user_id    UUID             NOT NULL REFERENCES users (id),
    token      TEXT             NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ      NOT NULL,
    created_at TIMESTAMPTZ      NOT NULL
);

CREATE TABLE change_email_requests
(
    id         UUID PRIMARY KEY NOT NULL,
    user_id    UUID             NOT NULL REFERENCES users (id),
    new_email  VARCHAR(255)     NOT NULL,
    token      TEXT             NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ      NOT NULL,
    created_at TIMESTAMPTZ      NOT NULL
);

CREATE TABLE oauth_user_accounts
(
    id               uuid PRIMARY KEY NOT NULL,
    user_id          uuid             NOT NULL REFERENCES users (id),
    -- see /oauth/provider.go for providers enum
    provider         TEXT             NOT NULL,
    provider_user_id text             NOT NULL,
    created_at       TIMESTAMPTZ      NOT NULL,

    UNIQUE (provider, provider_user_id)
);

CREATE TABLE files
(
    id           uuid PRIMARY KEY NOT NULL,
    owner_id     uuid             NOT NULL,
    name         text             NOT NULL,
    path         text             NOT NULL,
    preview_path text,
    hash         text             NOT NULL,
    type         text             NOT NULL,
    mime_type    text,
    purpose      text             NOT NULL,
    size         bigint           NOT NULL CHECK (size >= 0),
    created_at   timestamptz      NOT NULL,
    deleted_at   timestamptz,
    temp         boolean          NOT NULL DEFAULT false
);

CREATE TABLE entities
(
    id              uuid PRIMARY KEY NOT NULL,
    public_id       TEXT             NOT NULL,
    owner_id        UUID             NOT NULL REFERENCES users (id),
    preview_file_id uuid REFERENCES files (id),
    title           TEXT,
    description     TEXT,
    type            TEXT             NOT NULL,
    visibility      TEXT             NOT NULL,
    status          TEXT             NOT NULL,
    created_at      TIMESTAMPTZ      NOT NULL,
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX entities_public_id_type_uidx
    ON entities (public_id, type)
    WHERE deleted_at IS NULL;

CREATE TABLE books
(
    id            uuid PRIMARY KEY NOT NULL REFERENCES entities (id),
    author_name   TEXT             NOT NULL,
    author_link   TEXT,
    homepage      TEXT,
    release_date  TEXT             NOT NULL,
    cover_file_id uuid REFERENCES files (id)
);

CREATE TABLE entity_change_requests
(
    id          UUID PRIMARY KEY NOT NULL,
    entity_id   UUID             NOT NULL REFERENCES entities (id),
    user_id     UUID             NOT NULL REFERENCES users (id),
    diff        JSONB            NOT NULL,
    message     TEXT,
    status      TEXT             NOT NULL,
    revision    INT,

    reviewer_id UUID REFERENCES users (id),
    reviewed_at TIMESTAMPTZ,
    review_note TEXT,

    created_at  TIMESTAMPTZ      NOT NULL,
    updated_at  TIMESTAMPTZ
);

CREATE UNIQUE INDEX uidx_entity_change_requests_one_pending
    ON entity_change_requests (entity_id, user_id)
    WHERE status = 'pending';

CREATE TABLE event_logs
(
    id               UUID PRIMARY KEY,
    user_id          UUID REFERENCES users (id),
    type             TEXT        NOT NULL,
    entity_id        UUID REFERENCES entities (id),
    entity_change_id UUID REFERENCES entity_change_requests (id),
    message          TEXT,
    is_public        BOOLEAN,
    meta             JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE topics
(
    id          uuid PRIMARY KEY NOT NULL,
    type        TEXT             NOT NULL,
    public_id   TEXT             NOT NULL,
    name        TEXT             NOT NULL,
    description TEXT,
    created_at  TIMESTAMPTZ      NOT NULL,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,

    CONSTRAINT topics_entity_type_slug_uniq UNIQUE (type, public_id)
);

CREATE INDEX topics_entity_type_idx ON topics (type);
CREATE INDEX topics_deleted_at_idx ON topics (deleted_at);


CREATE TABLE entity_topics
(
    entity_id uuid NOT NULL REFERENCES entities (id),
    topic_id  uuid NOT NULL REFERENCES topics (id),

    CONSTRAINT entity_topics_pk PRIMARY KEY (entity_id, topic_id)
);

