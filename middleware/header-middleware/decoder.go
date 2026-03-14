package headermiddleware

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"sync"
)

type fieldDecoder struct {
	index int
	name  string
	kind  reflect.Kind
}

var decoderCache sync.Map // map[reflect.Type][]fieldDecoder

func DecodeHeaders(h http.Header, out any) error {
	v := reflect.ValueOf(out).Elem()
	t := v.Type()

	fields := getFieldInfo(t)

	for _, f := range fields {
		val := h.Get(f.name)
		if val == "" {
			continue
		}

		field := v.Field(f.index)

		switch f.kind {
		case reflect.String:
			field.SetString(val)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return fmt.Errorf("field %q: %w", f.name, err)
			}
			field.SetInt(i)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("field %q: %w", f.name, err)
			}
			field.SetUint(u)
		case reflect.Float32, reflect.Float64:
			fl, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("field %q: %w", f.name, err)
			}
			field.SetFloat(fl)
		case reflect.Bool:
			b, err := strconv.ParseBool(val)
			if err != nil {
				return fmt.Errorf("field %q: %w", f.name, err)
			}
			field.SetBool(b)
		}
	}

	return nil
}

func getFieldInfo(t reflect.Type) []fieldDecoder {
	if v, ok := decoderCache.Load(t); ok {
		return v.([]fieldDecoder)
	}

	fields := buildFieldInfo(t)

	actual, _ := decoderCache.LoadOrStore(t, fields)
	return actual.([]fieldDecoder)
}
func buildFieldInfo(t reflect.Type) []fieldDecoder {
	var fields []fieldDecoder

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		tag := f.Tag.Get("header")
		if tag == "" || tag == "-" {
			continue
		}

		fields = append(fields, fieldDecoder{
			index: i,
			name:  tag,
			kind:  f.Type.Kind(),
		})
	}

	return fields
}
