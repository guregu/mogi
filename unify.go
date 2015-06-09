package mogi

import (
	"database/sql/driver"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

// unify converts values to fit driver.Value,
// except []byte which is converted to string.
func unify(v interface{}) interface{} {
	// happy path
	switch x := v.(type) {
	case nil:
		return x
	case bool:
		return x
	case driver.Valuer:
		v, err := x.Value()
		if err != nil {
			panic(err)
		}
		return v

	// int64
	case int64:
		return x
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int16:
		return int64(x)
	case int8:
		return int64(x)
	case byte:
		return int64(x)

	// float64
	case float64:
		return x
	case float32:
		return float64(x)

	// string
	case string:
		return x
	case []byte:
		return string(x)

	// time.Time
	case time.Time:
		return x
	case *time.Time:
		if x == nil {
			return nil
		}
		return *x
	}

	// sad path
	rv := reflect.ValueOf(v)
	return reflectUnify(rv)
}

func reflectUnify(rv reflect.Value) interface{} {
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return nil
		}
		return reflectUnify(rv.Elem())
	case reflect.Bool:
		return rv.Bool()
	case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16, reflect.Int8:
		return rv.Int()
	case reflect.Float64, reflect.Float32:
		return rv.Float()
	case reflect.String:
		return rv.String()
	case reflect.Slice:
		if rv.Elem().Kind() == reflect.Int8 {
			return string(rv.Bytes())
		}
	}

	panic("couldn't unify value of type " + rv.Type().Name())
}

func unifyValues(values []driver.Value) []driver.Value {
	for i, v := range values {
		values[i] = unify(v)
	}
	return values
}

func unifyInterfaces(slice []interface{}) []interface{} {
	for i, v := range slice {
		slice[i] = unify(v)
	}
	return slice
}

func stringify(v interface{}) string {
	return fmt.Sprintf("%s", v)
}

func lowercase(strs []string) []string {
	lower := make([]string, 0, len(strs))
	for _, str := range strs {
		lower = append(lower, strings.ToLower(str))
	}
	return lower
}

func equals(src interface{}, to interface{}) bool {
	switch tox := to.(type) {
	case time.Time:
		// we need to convert source timestamps to time.Time
		if timeLayout == "" {
			break
		}
		var other time.Time
		switch srcx := src.(type) {
		case string:
			var err error
			if other, err = time.Parse(timeLayout, srcx); err != nil {
				goto deep
			}
		case []byte:
			var err error
			if other, err = time.Parse(timeLayout, string(srcx)); err != nil {
				goto deep
			}
		case time.Time:
			other = srcx
		}
		return tox.Format(timeLayout) == other.Format(timeLayout)
	case bool:
		// some drivers send booleans as 0 and 1
		switch srcx := src.(type) {
		case int64:
			return tox == (srcx != 0)
		case bool:
			return tox == srcx
		case string:
			other, ok := str2bool(srcx)
			if !ok {
				goto deep
			}
			return tox == other
		case []byte:
			other, ok := str2bool(string(srcx))
			if !ok {
				goto deep
			}
			return tox == other
		}
	}
deep:
	return reflect.DeepEqual(src, to)
}

// converts boolean-like strings to a bool
func str2bool(str string) (v bool, ok bool) {
	switch str {
	case "true", "1":
		return true, true
	case "false", "0":
		return false, true
	default:
		log.Println("mogi: unknown boolean string:", str)
		return false, false
	}
}
