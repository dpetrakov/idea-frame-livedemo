-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    login VARCHAR(32) NOT NULL,
    display_name VARCHAR(32) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create unique index on login
CREATE UNIQUE INDEX idx_users_login ON users(login);

-- Create index on created_at for sorting
CREATE INDEX idx_users_created_at ON users(created_at);

-- Add constraints
ALTER TABLE users
    ADD CONSTRAINT chk_login_length CHECK (char_length(login) >= 3 AND char_length(login) <= 32),
    ADD CONSTRAINT chk_display_name_length CHECK (char_length(display_name) >= 1 AND char_length(display_name) <= 32);

-- Comment on table
COMMENT ON TABLE users IS 'Пользователи системы с упрощённой регистрацией';
COMMENT ON COLUMN users.login IS 'Уникальный логин пользователя';
COMMENT ON COLUMN users.display_name IS 'Отображаемое имя';
COMMENT ON COLUMN users.password_hash IS 'Хэш пароля (bcrypt)';