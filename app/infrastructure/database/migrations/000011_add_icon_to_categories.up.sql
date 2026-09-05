-- Имя иконки (Lucide, в формате фронтенда, например i-lucide-cpu) —
-- та же конвенция, что и у user_links.icon.
ALTER TABLE categories ADD COLUMN icon VARCHAR(100) NOT NULL DEFAULT 'i-lucide-hash';

UPDATE categories SET icon = 'i-lucide-cpu' WHERE slug = 'tech';
UPDATE categories SET icon = 'i-lucide-users' WHERE slug = 'society';
UPDATE categories SET icon = 'i-lucide-palette' WHERE slug = 'culture';
UPDATE categories SET icon = 'i-lucide-flask-conical' WHERE slug = 'science';
UPDATE categories SET icon = 'i-lucide-briefcase' WHERE slug = 'business';
UPDATE categories SET icon = 'i-lucide-message-square-quote' WHERE slug = 'opinion';
