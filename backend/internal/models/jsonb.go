package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// JSONB is a PostgreSQL jsonb column backed by raw JSON bytes.
//
// It exists so the project can persist semi-structured payloads (the extracted
// AI intent, per-product score breakdowns) without pulling in gorm.io/datatypes
// for one type. Storing these as jsonb rather than normalised columns is a
// deliberate call: they are write-once audit/debug artefacts that we read back
// whole, and their shape is expected to churn as the ranker is tuned.
type JSONB []byte

// Value implements driver.Valuer.
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	if !json.Valid(j) {
		return nil, errors.New("models: JSONB holds invalid JSON")
	}
	return string(j), nil
}

// Scan implements sql.Scanner.
func (j *JSONB) Scan(src any) error {
	if src == nil {
		*j = nil
		return nil
	}
	switch v := src.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
		return nil
	case string:
		*j = append((*j)[0:0], v...)
		return nil
	default:
		return fmt.Errorf("models: cannot scan %T into JSONB", src)
	}
}

// MarshalJSON lets JSONB pass through encoding/json untouched instead of being
// base64-encoded as a byte slice would be.
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// UnmarshalJSON stores the raw bytes verbatim.
func (j *JSONB) UnmarshalJSON(b []byte) error {
	if j == nil {
		return errors.New("models: UnmarshalJSON on nil JSONB")
	}
	*j = append((*j)[0:0], b...)
	return nil
}

// GormDataType tells GORM which column type to create.
func (JSONB) GormDataType() string { return "jsonb" }

// ToJSONB marshals any value into a JSONB column value. On failure it returns
// SQL NULL rather than an error: none of the callers would meaningfully recover
// from a failed debug-payload marshal, and dropping a recommendation because
// its score breakdown would not serialise is the wrong trade.
func ToJSONB(v any) JSONB {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return JSONB(b)
}

// Decode unmarshals a JSONB column into dest.
//
// It is the read-side counterpart to ToJSONB, and it tolerates both an absent
// column and a payload whose shape has drifted since it was written. That
// tolerance is the point: these columns hold debug artefacts written by an
// earlier version of the ranker, and a decode failure should cost a reader some
// detail, never their request.
func (j JSONB) Decode(dest any) error {
	if len(j) == 0 {
		return nil
	}
	return json.Unmarshal(j, dest)
}
