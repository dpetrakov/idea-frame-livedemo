-- Откат миграции создания таблицы users

-- Удаление триггера
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Удаление функции
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление таблицы (каскадно удалит индексы и ограничения)
DROP TABLE IF EXISTS users;