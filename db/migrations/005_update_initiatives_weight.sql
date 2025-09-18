-- Migration: Update initiatives weight calculation to dynamic denominator
-- TK-003/DV: Корректная формула веса: учитываем только присутствующие атрибуты
-- weight = round((0.5*value + 0.3*speed - 0.2*cost) / (sum of used coeffs), 2)
-- Если все NULL — weight = 0.0

BEGIN;

CREATE OR REPLACE FUNCTION calculate_initiative_weight()
RETURNS TRIGGER AS $$
DECLARE
    denom NUMERIC := 0.0;
    num   NUMERIC := 0.0;
BEGIN
    IF NEW.value IS NOT NULL THEN
        num := num + (0.5 * NEW.value);
        denom := denom + 0.5;
    END IF;

    IF NEW.speed IS NOT NULL THEN
        num := num + (0.3 * NEW.speed);
        denom := denom + 0.3;
    END IF;

    IF NEW.cost IS NOT NULL THEN
        num := num - (0.2 * NEW.cost);
        denom := denom + 0.2;
    END IF;

    IF denom = 0 THEN
        NEW.weight := 0.0;
    ELSE
        NEW.weight := ROUND(num / denom, 2);
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

COMMIT;