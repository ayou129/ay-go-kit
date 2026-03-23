package dbx

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONBObject 用于 PostgreSQL JSONB 对象字段的自动序列化/反序列化。
// 适用场景：存储 JSON 对象，如 {"color": "red", "size": "L"}。
type JSONBObject map[string]any

func (j JSONBObject) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONBObject) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("dbx.JSONBObject: expected []byte")
	}
	result := make(JSONBObject)
	if err := json.Unmarshal(b, &result); err != nil {
		return err
	}
	*j = result
	return nil
}

// JSONBArray 用于 PostgreSQL JSONB 数组字段的自动序列化/反序列化。
// 适用场景：存储任意类型的 JSON 数组，如 [1, 2, 3] 或 ["a", "b"]。
type JSONBArray []any

func (j JSONBArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

func (j *JSONBArray) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("dbx.JSONBArray: expected []byte")
	}
	var result []any
	if err := json.Unmarshal(b, &result); err != nil {
		return err
	}
	*j = result
	return nil
}

// JSONBArrayStr 用于 PostgreSQL JSONB 字符串数组字段的自动序列化/反序列化。
// 适用场景：存储字符串数组，如 ["选项A", "选项B"]。
// 相比 JSONBArray，类型更明确，使用时不需要类型断言。
type JSONBArrayStr []string

func (s JSONBArrayStr) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *JSONBArrayStr) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("dbx.JSONBArrayStr: expected []byte")
	}
	var result []string
	if err := json.Unmarshal(b, &result); err != nil {
		return err
	}
	*s = result
	return nil
}
