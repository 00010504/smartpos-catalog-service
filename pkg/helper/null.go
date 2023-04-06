package helper

import (
	"database/sql"
)

func NullString(s string) (ns sql.NullString) {
	if s != "" {
		ns.String = s
		ns.Valid = true
	}
	return ns
}

func NullDouble(s float64) (ns sql.NullFloat64) {
	if s != 0 {
		ns.Float64 = s
		ns.Valid = true
	}
	return ns
}

func NullInt(s int64) (ns sql.NullInt64) {
	if s != 0 {
		ns.Int64 = s
		ns.Valid = true
	}
	return ns
}

func StringValue(ns sql.NullString) *string {
	if ns.Valid {
		s := ns.String
		return &s
	}
	return nil
}

func DoubleValue(ns sql.NullFloat64) *float64 {
	if ns.Valid {
		s := ns.Float64
		return &s
	}
	return nil
}

func Int64Value(ns sql.NullInt64) *int64 {
	if ns.Valid {
		s := ns.Int64
		return &s
	}
	return nil
}
