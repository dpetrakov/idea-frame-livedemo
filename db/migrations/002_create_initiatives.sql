-- Migration: Create initiatives table
-- TK-002/DV: Миграции для таблицы initiatives согласно db/schema.dbml
-- Поддерживает: создание инициатив, атрибуты оценки, вычисляемый вес
--
-- Формула веса: round((0.5*value + 0.3*speed - 0.2*cost) / 1.0, 2)
-- При NULL значениях атрибутов вес = 0

BEGIN;

-- Создание таблицы initiatives
CREATE TABLE IF NOT EXISTS initiatives (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title             VARCHAR(140) NOT NULL,
    description       TEXT,
    author_id         UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
    assignee_id       UUID REFERENCES users(id) ON DELETE SET NULL ON UPDATE RESTRICT,
    is_deleted        BOOLEAN NOT NULL DEFAULT FALSE,
    value             SMALLINT,
    speed             SMALLINT,
    cost              SMALLINT,
    weight            NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- CHECK ограничения для обеспечения целостности данных
ALTER TABLE initiatives 
    ADD CONSTRAINT chk_initiatives_title_length 
    CHECK (LENGTH(title) >= 1 AND LENGTH(title) <= 140);

ALTER TABLE initiatives 
    ADD CONSTRAINT chk_initiatives_description_length 
    CHECK (description IS NULL OR LENGTH(description) <= 10000);

ALTER TABLE initiatives 
    ADD CONSTRAINT chk_initiatives_value_range 
    CHECK (value IS NULL OR (value >= 1 AND value <= 5));

ALTER TABLE initiatives 
    ADD CONSTRAINT chk_initiatives_speed_range 
    CHECK (speed IS NULL OR (speed >= 1 AND speed <= 5));

ALTER TABLE initiatives 
    ADD CONSTRAINT chk_initiatives_cost_range 
    CHECK (cost IS NULL OR (cost >= 1 AND cost <= 5));

-- Индексы для оптимизации запросов (согласно schema.dbml)
CREATE INDEX idx_initiatives_weight_created ON initiatives (weight DESC, created_at DESC);
CREATE INDEX idx_initiatives_author ON initiatives (author_id);
CREATE INDEX idx_initiatives_assignee ON initiatives (assignee_id);
CREATE INDEX idx_initiatives_created_at ON initiatives (created_at);
CREATE INDEX idx_initiatives_is_deleted ON initiatives (is_deleted);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического обновления updated_at при изменении записи
CREATE TRIGGER update_initiatives_updated_at 
    BEFORE UPDATE ON initiatives 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Функция для автоматического пересчета веса
CREATE OR REPLACE FUNCTION calculate_initiative_weight()
RETURNS TRIGGER AS $$
BEGIN
    -- Если все атрибуты NULL, то вес = 0.0
    IF NEW.value IS NULL AND NEW.speed IS NULL AND NEW.cost IS NULL THEN
        NEW.weight = 0.0;
    -- Если есть хотя бы один атрибут, применяем формулу (NULL = 0 в расчетах)
    ELSE
        NEW.weight = ROUND(
            (0.5 * COALESCE(NEW.value, 0) + 
             0.3 * COALESCE(NEW.speed, 0) - 
             0.2 * COALESCE(NEW.cost, 0)) / 1.0, 2
        );
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггер для автоматического пересчета веса при создании/обновлении
CREATE TRIGGER calculate_initiatives_weight
    BEFORE INSERT OR UPDATE OF value, speed, cost ON initiatives
    FOR EACH ROW 
    EXECUTE FUNCTION calculate_initiative_weight();

COMMIT;