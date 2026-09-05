package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONArray stores any slice as a Postgres jsonb column. It exists so the
// project can keep structured columns (article tags, content blocks)
// without pulling in gorm.io/datatypes as an extra dependency.
type JSONArray[T any] []T

func (j JSONArray[T]) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal([]T(j))
}

func (j *JSONArray[T]) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("models.JSONArray: unsupported Scan source type")
	}

	var result []T
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	*j = result
	return nil
}

func (JSONArray[T]) GormDataType() string {
	return "jsonb"
}
