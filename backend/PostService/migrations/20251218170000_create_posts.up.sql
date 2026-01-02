CREATE TABLE IF NOT EXISTS posts (
    post_id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_posts_author_created_id
    ON posts (author_id, created_at DESC, post_id DESC);

