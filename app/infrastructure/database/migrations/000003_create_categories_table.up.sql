CREATE TABLE categories (
    slug VARCHAR(50) PRIMARY KEY,
    title VARCHAR(255) NOT NULL
);

INSERT INTO categories (slug, title) VALUES
    ('tech', 'Технологии'),
    ('society', 'Общество'),
    ('culture', 'Культура'),
    ('science', 'Наука'),
    ('business', 'Бизнес'),
    ('opinion', 'Мнения');
