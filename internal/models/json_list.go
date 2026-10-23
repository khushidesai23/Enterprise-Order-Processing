package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSONList stores a slice in a single jsonb column. A nil list is stored as
// an empty array so readers never see null.
type JSONList[T any] []T

func (l JSONList[T]) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}

	raw, err := json.Marshal([]T(l))
	if err != nil {
		return nil, err
	}

	return string(raw), nil
}

func (l *JSONList[T]) Scan(value any) error {
	var raw []byte

	switch v := value.(type) {
	case nil:
		*l = JSONList[T]{}
		return nil
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("JSONList: unsupported type %T", value)
	}

	var items []T
	if err := json.Unmarshal(raw, &items); err != nil {
		return err
	}

	*l = items

	return nil
}

// GormDataType makes AutoMigrate create the column as jsonb.
func (JSONList[T]) GormDataType() string {
	return "jsonb"
}
