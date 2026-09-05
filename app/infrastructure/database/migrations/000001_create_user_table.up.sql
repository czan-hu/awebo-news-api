CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    role VARCHAR(50) NOT NULL DEFAULT 'user',
    title VARCHAR(255) NOT NULL DEFAULT '',

    name VARCHAR(255) NOT NULL DEFAULT '',
    username VARCHAR(255) UNIQUE,
    email VARCHAR(255) UNIQUE NOT NULL,

    avatar_path VARCHAR(255) NOT NULL DEFAULT '',
    bio TEXT NOT NULL DEFAULT '',
    location VARCHAR(255) NOT NULL DEFAULT '',

    verification_code VARCHAR(255),
    verification_code_expires_at TIMESTAMPTZ,
    code_requested_at TIMESTAMPTZ,

    profile_completed BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_username ON users (username) WHERE username IS NOT NULL;
