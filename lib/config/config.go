package config

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func Load(configGroup any) { load("", configGroup) }

func load(prefix string, configGroup any) {
	v := reflect.ValueOf(configGroup)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		panic("config group must be a pointer to a struct")
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)
		configKey := prefix + camelToSnake(field.Name)

		// Handle pointer-to-struct fields
		if fieldValue.Kind() == reflect.Ptr && fieldValue.Type().Elem().Kind() == reflect.Struct {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			}
			load(configKey+".", fieldValue.Interface())
			continue
		}

		// Handle inline struct fields
		if fieldValue.Kind() == reflect.Struct && fieldValue.Type() != reflect.TypeOf(time.Duration(0)) {
			nested := fieldValue.Addr().Interface()
			load(configKey+".", nested)
			continue
		}

		if !fieldValue.CanSet() {
			continue
		}

		// Special-case exact type time.Duration
		if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {
			// Let viper parse durations like "30s", "2m", etc.
			d := viper.GetDuration(configKey)
			// If empty string in config, GetDuration returns 0; that may be intended.
			fieldValue.SetInt(int64(d))
			continue
		}

		switch fieldValue.Kind() {
		case reflect.Int:
			fieldValue.SetInt(int64(viper.GetInt(configKey)))
		case reflect.Int64:
			fieldValue.SetInt(int64(viper.GetInt64(configKey)))
		case reflect.String:
			fieldValue.SetString(viper.GetString(configKey))
		case reflect.Bool:
			fieldValue.SetBool(viper.GetBool(configKey))
		case reflect.Slice:
			elem := fieldValue.Type().Elem()

			// []time.Duration
			if elem == reflect.TypeOf(time.Duration(0)) {
				ss := viper.GetStringSlice(configKey) // expect ["5s","1m","250ms"]
				out := reflect.MakeSlice(fieldValue.Type(), 0, len(ss))
				for _, s := range ss {
					d, err := time.ParseDuration(s)
					if err != nil {
						panic(fmt.Sprintf("invalid duration in %s: %q: %v", configKey, s, err))
					}
					out = reflect.Append(out, reflect.ValueOf(d))
				}
				fieldValue.Set(out)
				break
			}

			// Primitive slices
			switch elem.Kind() {
			case reflect.String:
				fieldValue.Set(reflect.ValueOf(viper.GetStringSlice(configKey)))
			case reflect.Int:
				ints := viper.GetIntSlice(configKey)
				out := reflect.MakeSlice(fieldValue.Type(), len(ints), len(ints))
				for i, n := range ints {
					out.Index(i).SetInt(int64(n))
				}
				fieldValue.Set(out)
			case reflect.Int64:
				// viper has no GetInt64Slice; convert from []int or generic []any
				raw := viper.Get(configKey)
				switch vv := raw.(type) {
				case []int:
					out := reflect.MakeSlice(fieldValue.Type(), len(vv), len(vv))
					for i, n := range vv {
						out.Index(i).SetInt(int64(n))
					}
					fieldValue.Set(out)
				case []any:
					out := reflect.MakeSlice(fieldValue.Type(), len(vv), len(vv))
					for i, x := range vv {
						out.Index(i).SetInt(toInt64(x))
					}
					fieldValue.Set(out)
				default:
					// try UnmarshalKey as a fallback
					assignViaUnmarshal(configKey, fieldValue)
				}
			case reflect.Bool:
				// viper lacks GetBoolSlice; convert generically
				raw := viper.Get(configKey)
				switch vv := raw.(type) {
				case []bool:
					fieldValue.Set(reflect.ValueOf(vv))
				case []any:
					out := reflect.MakeSlice(fieldValue.Type(), len(vv), len(vv))
					for i, x := range vv {
						out.Index(i).SetBool(toBool(x))
					}
					fieldValue.Set(out)
				default:
					assignViaUnmarshal(configKey, fieldValue)
				}
			case reflect.Struct, reflect.Map, reflect.Interface:
				// Complex slice: delegate to UnmarshalKey (also handles nested structs)
				assignViaUnmarshal(configKey, fieldValue)
			default:
				assignViaUnmarshal(configKey, fieldValue)
			}
		default:
			// Fallback: let viper/mapstructure try (covers maps, nested types, etc.)
			assignViaUnmarshal(configKey, fieldValue)
		}
	}

	// Optional: post-load defaults hook
	if c, ok := configGroup.(interface{ FillDefaults() }); ok {
		c.FillDefaults()
	}
}

func camelToSnake(str string) string {
	var b strings.Builder
	b.Grow(len(str) * 2)
	for i, r := range str {
		if unicode.IsUpper(r) && i != 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func toInt64(x any) int64 {
	switch n := x.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case float32:
		return int64(n)
	case float64:
		return int64(n)
	case string:
		d, err := time.ParseDuration(n) // allow "5s" for []int64 durations? probably not—fallback parse int
		if err == nil {
			return int64(d)
		}
		// last resort: parse as integer
		var v int64
		_, _ = fmt.Sscan(n, &v)
		return v
	default:
		return 0
	}
}

func toBool(x any) bool {
	switch b := x.(type) {
	case bool:
		return b
	case string:
		switch strings.ToLower(b) {
		case "1", "t", "true", "yes", "y", "on":
			return true
		}
		return false
	default:
		return false
	}
}

func assignViaUnmarshal(key string, dest reflect.Value) {
	ptr := reflect.New(dest.Type()).Interface()

	// duration-aware decode hook
	hook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
	)

	if err := viper.UnmarshalKey(key, ptr, viper.DecodeHook(hook)); err != nil {
		panic(fmt.Sprintf("unmarshal %q failed: %v", key, err))
	}
	dest.Set(reflect.ValueOf(ptr).Elem())
}
