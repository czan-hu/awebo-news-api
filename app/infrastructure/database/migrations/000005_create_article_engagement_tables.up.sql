CREATE TABLE article_likes (
    article_id INTEGER NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (article_id, user_id)
);

CREATE TABLE article_saves (
    article_id INTEGER NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (article_id, user_id)
);

CREATE TABLE article_views (
    article_id INTEGER NOT NULL REFERENCES articles (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (article_id, user_id)
);

CREATE INDEX idx_article_likes_user_id ON article_likes (user_id);
CREATE INDEX idx_article_saves_user_id ON article_saves (user_id);
CREATE INDEX idx_article_views_user_id ON article_views (user_id, viewed_at DESC);
