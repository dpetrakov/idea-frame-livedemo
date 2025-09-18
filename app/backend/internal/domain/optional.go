package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// OptionalUUID представляет три состояния поля UUID в JSON: отсутствует / null / значение
// Present = true, Null = true  -> поле присутствовало и было null
// Present = true, Null = false -> поле присутствовало и содержало значение (Valid указывает парсинг UUID)
// Present = false              -> поле отсутствовало в JSON
// Если строка была передана, но формат UUID неверен — Valid = false (решение об ошибке на уровне сервиса)
type OptionalUUID struct {
	Present bool
	Null    bool
	Valid   bool
	Value   uuid.UUID
}

func (o *OptionalUUID) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	o.Present = true
	// null
	if s == "null" {
		o.Null = true
		o.Valid = true
		return nil
	}
	// строка
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		str, err := unquote(s)
		if err != nil {
			// некорректная JSON-строка
			return fmt.Errorf("invalid string for uuid: %w", err)
		}
		u, err := uuid.Parse(str)
		if err != nil {
			// оставляем Present=true, Null=false, Valid=false — обработаем в сервисе
			o.Null = false
			o.Valid = false
			return nil
		}
		o.Null = false
		o.Valid = true
		o.Value = u
		return nil
	}
	// любой другой тип — невалидный JSON по контракту
	return fmt.Errorf("invalid type for uuid field: %s", s)
}

func unquote(s string) (string, error) {
	var v string
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return "", err
	}
	return v, nil
}