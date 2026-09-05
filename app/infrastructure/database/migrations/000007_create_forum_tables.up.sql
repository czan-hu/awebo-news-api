CREATE TABLE forum_topics (
    id VARCHAR(255) PRIMARY KEY,

    title VARCHAR(500) NOT NULL,
    category VARCHAR(100) NOT NULL,
    author_id UUID NOT NULL REFERENCES users (id),
    body TEXT NOT NULL,

    pinned BOOLEAN NOT NULL DEFAULT FALSE,
    comment_count BIGINT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE forum_comments (
    id VARCHAR(255) PRIMARY KEY,

    topic_id VARCHAR(255) NOT NULL REFERENCES forum_topics (id) ON DELETE CASCADE,
    parent_id VARCHAR(255) REFERENCES forum_comments (id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users (id),
    body TEXT NOT NULL,
    score INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE forum_votes (
    comment_id VARCHAR(255) NOT NULL REFERENCES forum_comments (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    direction SMALLINT NOT NULL,
    PRIMARY KEY (comment_id, user_id)
);

CREATE INDEX idx_forum_topics_category ON forum_topics (category);
CREATE INDEX idx_forum_topics_pinned_created_at ON forum_topics (pinned DESC, created_at DESC);
CREATE INDEX idx_forum_comments_topic_id ON forum_comments (topic_id);
CREATE INDEX idx_forum_comments_parent_id ON forum_comments (parent_id);
