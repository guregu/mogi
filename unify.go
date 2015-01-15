package mogi

import (
	"database/sql/driver"
	"reflect"
	"time"
)

/*
int64
float64
bool
[]byte
string   [*] everywhere except from Rows.Next.
time.Time
*/

// unify converts values to fit driver.Value,
// except []byte which is converted to string.
func unify(v interface{}) interface{} {
	// happy path
	switch x := v.(type) {
	case nil:
		return x
	case bool:
		return x

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
