package sqlutils

import (
	"database/sql"
	"errors"
	"reflect"
)

func ScanRow[T any](row *sql.Row) (*T, error) {
	var result T
	fields := structFields(&result)
	if err := row.Scan(fields...); err != nil {
		return nil, err
	}
	return &result, nil
}

func ScanRows[T any](rows *sql.Rows) ([]T, error) {
	var results []T
	for rows.Next() {
		var item T
		fields := structFields(&item)
		if err := rows.Scan(fields...); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func structFields(s any) []any {
	v := reflect.ValueOf(s).Elem()
	fields := make([]any, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		fields[i] = v.Field(i).Addr().Interface()
	}
	return fields
}

// ScanOptional scans a single row, returning nil (not error) if no rows found
func ScanOptional[T any](row *sql.Row) (*T, error) {
	result, err := ScanRow[T](row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return result, err
}
