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


CREATE TABLE user_activity_logs
(
    id          UUID PRIMARY KEY NOT NULL,
    user_id     UUID REFERENCES users (id),
    action_type INT              NOT NULL,
    is_public   BOOLEAN          NOT NULL DEFAULT FALSE,
    entity_type VARCHAR(100),
    entity_id   BIGINT,
    meta        JSONB,
    created_at  TIMESTAMPTZ      NOT NULL
);

CREATE INDEX idx_user_activity_logs_user_id ON user_activity_logs (user_id);
CREATE INDEX idx_user_activity_logs_action_type ON user_activity_logs (action_type);
CREATE INDEX idx_user_activity_logs_public_feed ON user_activity_logs (created_at DESC) WHERE is_public = TRUE;

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
    public_id       TEXT  UNIQUE     NOT NULL,
    owner_id        UUID             NOT NULL,
    preview_file_id uuid REFERENCES files (id),
    title           TEXT,
    type            TEXT             NOT NULL,
    visibility      TEXT             NOT NULL,
    status          TEXT             NOT NULL,
    created_at      TIMESTAMPTZ      NOT NULL,
    updated_at      TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ
);

CREATE TABLE entity_change_logs
(
    id         uuid PRIMARY KEY         NOT NULL,
    entity_id  UUID                     NOT NULL REFERENCES entities (id) ON DELETE CASCADE,
    user_id    UUID                     NOT NULL,
    action     TEXT                     NOT NULL,
    metadata   JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE books
(
    id            uuid PRIMARY KEY NOT NULL REFERENCES entities (id),
    description   TEXT             NOT NULL,
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
