package models

import "encoding/json"

// JSON is a custom type for handling JSONB in PostgreSQL
type JSON []byte

// MarshalJSON returns j as the JSON encoding of j.
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON sets *j to a copy of data.
func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return nil
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Value returns j as a value. This does a validating unmarshal into another
// RawMessage. If j is invalid json, it will return an error.
func (j JSON) Value() (interface{}, error) {
	var m interface{}
	var err = json.Unmarshal(j, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Scan implements the sql.Scanner interface.
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return nil
	}
	*j = append((*j)[0:0], s...)
	return nil
}
