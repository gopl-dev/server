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

CREATE TABLE user_sessions
(
    id         UUID PRIMARY KEY NOT NULL,
    user_id    BIGINT                NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ           NOT NULL,
    updated_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ           NOT NULL
);






