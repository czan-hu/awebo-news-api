CREATE TABLE articles (
    id SERIAL PRIMARY KEY,

    slug VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(500) NOT NULL,
    excerpt TEXT NOT NULL,
    category_slug VARCHAR(50) NOT NULL REFERENCES categories (slug),
    author_id UUID NOT NULL REFERENCES users (id),

    published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    reading_minutes INTEGER NOT NULL DEFAULT 1,
    cover VARCHAR(500) NOT NULL DEFAULT '',
    featured BOOLEAN NOT NULL DEFAULT FALSE,

    tags JSONB NOT NULL DEFAULT '[]',
    content JSONB NOT NULL DEFAULT '[]',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_articles_category_slug ON articles (category_slug);
CREATE INDEX idx_articles_author_id ON articles (author_id);
CREATE INDEX idx_articles_published_at ON articles (published_at DESC);
CREATE INDEX idx_articles_featured ON articles (featured);
CREATE INDEX idx_articles_tags ON articles USING GIN (tags);
