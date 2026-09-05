-- Демо-аккаунт администратора для локальной разработки — без него некому
-- рассматривать заявки на авторство. Войти можно обычным флоу
-- (запрос кода на admin@awebo.example); в MAIL_DRIVER=console код смотрите
-- в логах контейнера api.
INSERT INTO users (id, role, title, name, username, email, avatar_path, bio, location, profile_completed)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    'admin',
    'Администратор awebo',
    'Админ awebo',
    'admin',
    'admin@awebo.example',
    'https://api.dicebear.com/7.x/initials/svg?seed=Admin',
    '',
    '',
    TRUE
);
