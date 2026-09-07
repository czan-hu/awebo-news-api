-- Способы связи с редакцией — выводятся иконками в подвале сайта.
-- Иконка хранится «сырой» разметкой <svg>...</svg>, фронтенд вставляет её
-- как есть (см. SvgIcon.vue). Порядок задаётся sort_order (по возрастанию).
CREATE TABLE contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    label VARCHAR(255) NOT NULL,
    href VARCHAR(500) NOT NULL,
    icon TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Стартовый набор. Поправьте ссылки на свои, удалите лишние строки;
-- чтобы добавить новый способ связи — INSERT с бОльшим sort_order:
--   INSERT INTO contacts (label, href, icon, sort_order)
--   VALUES ('Дзен', 'https://dzen.ru/awebo', '<svg ...>...</svg>', 50);
INSERT INTO contacts (label, href, icon, sort_order) VALUES
    ('Telegram', 'https://t.me/awebo',
     '<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M9.78 18.65l.28-4.23 7.68-6.92c.34-.31-.07-.46-.52-.19L7.74 13.3 3.64 12c-.88-.25-.89-.86.2-1.3l15.97-6.16c.73-.33 1.43.18 1.15 1.3l-2.72 12.81c-.19.91-.74 1.13-1.5.71L12.6 16.3l-1.99 1.93c-.23.23-.42.42-.83.42Z"/></svg>',
     10),
    ('ВКонтакте', 'https://vk.com/awebo',
     '<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M13.94 17.4c-5.55 0-8.72-3.8-8.85-10.13h2.78c.09 4.65 2.14 6.62 3.76 7.02V7.27h2.62v4.01c1.6-.17 3.28-1.99 3.85-4.01h2.62c-.43 2.48-2.25 4.3-3.55 5.05 1.3.61 3.37 2.2 4.15 5.08h-2.88c-.61-1.9-2.15-3.37-4.14-3.57v3.57h-.32Z"/></svg>',
     20),
    ('X (Twitter)', 'https://x.com/awebo',
     '<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M18.24 2.25h3.3l-7.2 8.23L22.5 21.75h-6.63l-5.2-6.8-5.94 6.8H1.42l7.7-8.8L1.5 2.25h6.8l4.7 6.2 5.24-6.2Zm-1.16 17.52h1.83L7.02 4.12H5.06l12.02 15.65Z"/></svg>',
     30),
    ('Написать на почту', 'mailto:hello@awebo.news',
     '<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M2.5 6.5A2.5 2.5 0 0 1 5 4h14a2.5 2.5 0 0 1 2.5 2.5v11A2.5 2.5 0 0 1 19 20H5a2.5 2.5 0 0 1-2.5-2.5v-11Zm2.06-.5 7.44 5.57L19.44 6H4.56ZM20 7.68l-7.4 5.54a1 1 0 0 1-1.2 0L4 7.68V18h16V7.68Z"/></svg>',
     40);
