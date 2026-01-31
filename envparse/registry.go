package envparse

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v11"
)

const (
	EnvFileDefault = ".env.save"
)

var (
	reg = NewRegistry()
)

// Parse parses environment variables into the struct, registers them, and returns any error
func Parse(cfg any) error {
	if err := env.Parse(cfg); err != nil {
		return err
	}
	reg.register(cfg)
	return nil
}

func ToEnvFile(path string) error {
	out := reg.ToEnv()
	return os.WriteFile(path, []byte(out), os.ModePerm)
}

type EnvEntry struct {
	Key      string
	Value    any
	Default  string
	Required bool
}

type Registry struct {
	entries         map[string]EnvEntry
	registeredTypes map[reflect.Type]bool
}

func NewRegistry() *Registry {
	return &Registry{
		entries:         make(map[string]EnvEntry),
		registeredTypes: make(map[reflect.Type]bool),
	}
}

func (r *Registry) Add(key string, entry EnvEntry) {
	r.entries[key] = entry
}

func (r *Registry) Get(key string) (EnvEntry, bool) {
	e, ok := r.entries[key]
	return e, ok
}

func (r *Registry) All() map[string]EnvEntry {
	return r.entries
}

func (r *Registry) ToEnv() string {
	var sb strings.Builder
	for key, entry := range r.entries {
		sb.WriteString(fmt.Sprintf("%s=%v\n", key, entry.Value))
	}
	return sb.String()
}

func (r *Registry) register(s any) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()

	if r.registeredTypes[t] {
		return
	}
	r.registeredTypes[t] = true

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		// Parse env tag (handles "KEY,required" format)
		parts := strings.Split(envTag, ",")
		key := parts[0]
		required := len(parts) > 1 && parts[1] == "required"

		entry := EnvEntry{
			Key:      key,
			Value:    v.Field(i).Interface(),
			Default:  field.Tag.Get("envDefault"),
			Required: required,
		}

		r.Add(key, entry)
	}
}
