-- Создание таблицы пользователей для системы фрейминга инициатив
-- Согласно db/schema.dbml и PRD требованиям

-- Включение расширения для UUID (если не включено)
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Создание таблицы users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(32) NOT NULL,
    display_name VARCHAR(32) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создание уникального индекса для login
CREATE UNIQUE INDEX idx_users_login ON users (login);

-- Создание индекса для created_at
CREATE INDEX idx_users_created_at ON users (created_at);

-- Добавление CHECK ограничений для валидации согласно PRD
ALTER TABLE users ADD CONSTRAINT check_login_length 
    CHECK (char_length(login) >= 3 AND char_length(login) <= 32);

ALTER TABLE users ADD CONSTRAINT check_login_format 
    CHECK (login ~ '^[a-zA-Z0-9_-]+$');

ALTER TABLE users ADD CONSTRAINT check_display_name_length 
    CHECK (char_length(display_name) >= 1 AND char_length(display_name) <= 32);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();