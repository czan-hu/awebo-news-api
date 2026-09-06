ALTER TABLE categories ALTER COLUMN icon TYPE VARCHAR(100) USING left(icon, 100);
ALTER TABLE categories ALTER COLUMN icon SET DEFAULT 'i-lucide-hash';

UPDATE categories SET icon = 'i-lucide-cpu' WHERE slug = 'tech';
UPDATE categories SET icon = 'i-lucide-users' WHERE slug = 'society';
UPDATE categories SET icon = 'i-lucide-palette' WHERE slug = 'culture';
UPDATE categories SET icon = 'i-lucide-flask-conical' WHERE slug = 'science';
UPDATE categories SET icon = 'i-lucide-briefcase' WHERE slug = 'business';
UPDATE categories SET icon = 'i-lucide-message-square-quote' WHERE slug = 'opinion';
