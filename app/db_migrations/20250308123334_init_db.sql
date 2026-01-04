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


CREATE TABLE entities
(
    id         uuid PRIMARY KEY NOT NULL,
    owner_id   UUID             NOT NULL,
    title      TEXT,
    type       SMALLINT         NOT NULL DEFAULT 0, -- see ds.EntityType (0 = Draft)
    visibility SMALLINT         NOT NULL DEFAULT 0, -- Maps to ds.Visibility (0 = Public)
    status     SMALLINT         NOT NULL DEFAULT 0, -- Maps to ds.Status (0 = UnderReview)
    url_name   TEXT             NOT NULL,
    created_at TIMESTAMPTZ      NOT NULL,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
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
    id           uuid PRIMARY KEY NOT NULL REFERENCES entities (id),
    description  TEXT             NOT NULL,
    author_name  TEXT             NOT NULL,
    author_link  TEXT,
    homepage     TEXT,
    release_date TEXT             NOT NULL,
    cover_image  TEXT             NOT NULL
);

