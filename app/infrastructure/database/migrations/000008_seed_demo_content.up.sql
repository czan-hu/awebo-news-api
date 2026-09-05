-- Editorial account used as the author of seeded demo content, so freshly
-- provisioned environments aren't completely empty for the frontend to render.
INSERT INTO users (id, role, title, name, username, email, avatar_path, bio, location, profile_completed)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'editor',
    'Редактор раздела «Технологии»',
    'Марат Аширов',
    'marat',
    'marat@awebo.example',
    'https://api.dicebear.com/7.x/initials/svg?seed=Marat',
    'Пишу о том, как технологии меняют повседневность.',
    'Алматы',
    TRUE
);

INSERT INTO articles (slug, title, excerpt, category_slug, author_id, published_at, reading_minutes, cover, featured, tags, content)
VALUES (
    'tikhie-interfeysy-vozvraschenie-k-tekstu',
    'Тихие интерфейсы: возвращение к тексту',
    'Почему всё больше продуктов отказываются от ярких анимаций в пользу спокойного, текстового опыта.',
    'tech',
    '00000000-0000-0000-0000-000000000001',
    now(),
    7,
    'https://images.unsplash.com/photo-1517842645767-c639042777db',
    TRUE,
    '["дизайн", "интерфейсы", "текст"]',
    '[
        {"type":"p","text":"Последние несколько лет продуктовый дизайн двигался в сторону максимальной выразительности: анимации, звук, тактильный отклик."},
        {"type":"h2","text":"Что изменилось"},
        {"type":"p","text":"Сегодня всё больше команд возвращаются к простому тексту как основному носителю информации."},
        {"type":"quote","text":"Тихий интерфейс — это не отсутствие дизайна, а его высшая форма.","cite":"Дитер Рамс"},
        {"type":"ul","items":["Меньше анимаций","Больше читаемого текста","Явные состояния вместо намёков"]},
        {"type":"callout","emoji":"🪶","text":"Попробуйте убрать одну анимацию из своего продукта — и посмотрите, станет ли от этого хуже."}
    ]'
);

INSERT INTO forum_topics (id, title, category, author_id, body, pinned, comment_count)
VALUES (
    'tihie-interfeysy',
    'Тихие интерфейсы — это надолго или просто мода?',
    'Технологии',
    '00000000-0000-0000-0000-000000000001',
    'Заметил тренд на спокойные, текстовые интерфейсы без лишней анимации. Это осознанный отказ от избыточности или просто очередной цикл моды в дизайне?',
    TRUE,
    0
);
