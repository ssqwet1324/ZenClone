CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    login VARCHAR NOT NULL UNIQUE,
    password VARCHAR(100) NOT NULL,
    refresh_token VARCHAR,
    username VARCHAR NOT NULL UNIQUE,
    first_name VARCHAR NOT NULL,
    last_name VARCHAR,
    bio TEXT,
    avatar_url TEXT NOT NULL DEFAULT 'default',
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS subscriptions (
    follower_id UUID NOT NULL,
    following_id UUID NOT NULL,
    PRIMARY KEY (follower_id, following_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_subscriptions_following ON subscriptions(following_id);

CREATE INDEX idx_users_name_search ON users (first_name, last_name);

