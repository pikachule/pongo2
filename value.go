package pongo2

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Value struct {
	v reflect.Value
}

// Converts any given value to a pongo2.Value
// Usually being used within own functions passed to a template
// through a Context or within filter functions.
//
// Example:
//     AsValue("my string")
func AsValue(i interface{}) *Value {
	return &Value{
		v: reflect.ValueOf(i),
	}
}

func (v *Value) getResolvedValue() reflect.Value {
	if v.v.IsValid() && v.v.Kind() == reflect.Ptr {
		return v.v.Elem()
	}
	return v.v
}

// Checks whether the underlying value is a string
func (v *Value) IsString() bool {
	return v.getResolvedValue().Kind() == reflect.String
}

// Checks whether the underlying value is a bool
func (v *Value) IsBool() bool {
	return v.getResolvedValue().Kind() == reflect.Bool
}

// Checks whether the underlying value is a float
func (v *Value) IsFloat() bool {
	return v.getResolvedValue().Kind() == reflect.Float32 ||
		v.getResolvedValue().Kind() == reflect.Float64
}

// Checks whether the underlying value is an integer
func (v *Value) IsInteger() bool {
	return v.getResolvedValue().Kind() == reflect.Int ||
		v.getResolvedValue().Kind() == reflect.Int8 ||
		v.getResolvedValue().Kind() == reflect.Int16 ||
		v.getResolvedValue().Kind() == reflect.Int32 ||
		v.getResolvedValue().Kind() == reflect.Int64 ||
		v.getResolvedValue().Kind() == reflect.Uint ||
		v.getResolvedValue().Kind() == reflect.Uint8 ||
		v.getResolvedValue().Kind() == reflect.Uint16 ||
		v.getResolvedValue().Kind() == reflect.Uint32 ||
		v.getResolvedValue().Kind() == reflect.Uint64
}

// Checks whether the underlying value is either an integer
// or a float.
func (v *Value) IsNumber() bool {
	return v.IsInteger() || v.IsFloat()
}

// Checks whether the underlying value is NIL
func (v *Value) IsNil() bool {
	//fmt.Printf("%+v\n", v.getResolvedValue().Type().String())
	return !v.getResolvedValue().IsValid()
}

// Returns a string for the underlying value. If this value is not
// of type string, pongo2 tries to convert it. Currently the following
// types for underlying values are supported:
//
//     1. string
//     2. int/uint (any size)
//     3. float (any precision)
//     4. bool
//     5. time.Time
//
// NIL values will lead to an empty string. Unsupported types are leading
// to their respective type name.
func (v *Value) String() string {
	if v.IsNil() {
		return ""
	}

	switch v.getResolvedValue().Kind() {
	case reflect.String:
		return v.getResolvedValue().String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.getResolvedValue().Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.getResolvedValue().Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.getResolvedValue().Float())
	case reflect.Bool:
		if v.Bool() {
			return "True"
		} else {
			return "False"
		}
	case reflect.Struct:
		if v.getResolvedValue().Type() == reflect.TypeOf(time.Time{}) {
			return v.getResolvedValue().Interface().(time.Time).String()
		}
		if t, ok := v.Interface().(fmt.Stringer); ok {
			return t.String()
		}
	}

	logf("Value.String() not implemented for type: %s\n", v.getResolvedValue().Kind().String())
	return v.getResolvedValue().String()
}

// Returns the underlying value as an integer (converts the underlying
// value, if necessary). If it's not possible to convert the underlying value,
// it will return 0.
func (v *Value) Integer() int {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(v.getResolvedValue().Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int(v.getResolvedValue().Uint())
	case reflect.Float32, reflect.Float64:
		return int(v.getResolvedValue().Float())
	case reflect.String:
		// Try to convert from string to int (base 10)
		f, err := strconv.ParseFloat(v.getResolvedValue().String(), 64)
		if err != nil {
			return 0
		}
		return int(f)
	default:
		logf("Value.Integer() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return 0
	}
}

// Returns the underlying value as a float (converts the underlying
// value, if necessary). If it's not possible to convert the underlying value,
// it will return 0.0.
func (v *Value) Float() float64 {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.getResolvedValue().Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.getResolvedValue().Uint())
	case reflect.Float32, reflect.Float64:
		return v.getResolvedValue().Float()
	case reflect.String:
		// Try to convert from string to float64 (base 10)
		f, err := strconv.ParseFloat(v.getResolvedValue().String(), 64)
		if err != nil {
			return 0.0
		}
		return f
	default:
		logf("Value.Float() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return 0.0
	}
}

// Returns the underlying value as bool. If the value is not bool, false
// will always be returned. If you're looking for true/false-evaluation of the
// underlying value, have a look on the IsTrue()-function.
func (v *Value) Bool() bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Bool:
		return v.getResolvedValue().Bool()
	default:
		logf("Value.Bool() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return false
	}
}

// Tried to evaluate the underlying value the Pythonic-way:
//
// Returns TRUE in one the following cases:
//
//     * int != 0
//     * uint != 0
//     * float != 0.0
//     * len(array/chan/map/slice/string) > 0
//     * bool == true
//
// Otherwise returns always FALSE.
func (v *Value) IsTrue() bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.getResolvedValue().Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.getResolvedValue().Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.getResolvedValue().Float() != 0
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.getResolvedValue().Len() > 0
	case reflect.Bool:
		return v.getResolvedValue().Bool()
	default:
		logf("Value.IsTrue() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return false
	}
}

// Tries to negate the underlying value. It's mainly used for
// the NOT-operator and in conjunction with a call to
// return_value.IsTrue() afterwards.
//
// Example:
//     AsValue(1).Negate().IsTrue() == false
func (v *Value) Negate() *Value {
	switch v.getResolvedValue().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Integer() != 0 {
			return AsValue(0)
		} else {
			return AsValue(1)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() != 0.0 {
			return AsValue(float64(0.0))
		} else {
			return AsValue(float64(1.1))
		}
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return AsValue(v.getResolvedValue().Len() == 0)
	case reflect.Bool:
		return AsValue(!v.getResolvedValue().Bool())
	default:
		logf("Value.IsTrue() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return AsValue(true)
	}
}

// Returns the length for an array, chan, map, slice or string.
// Otherwise it will return 0.
func (v *Value) Len() int {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.getResolvedValue().Len()
	default:
		logf("Value.Len() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return 0
	}
}

// Slices an array, slice or string. Otherwise it will
// return an empty []int.
func (v *Value) Slice(i, j int) *Value {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		return AsValue(v.getResolvedValue().Slice(i, j).Interface())
	default:
		logf("Value.Slice() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return AsValue([]int{})
	}
}

// Get the i-th item of an array, slice or string. Otherwise
// it will return NIL.
func (v *Value) Index(i int) *Value {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Slice:
		if i >= v.Len() {
			return AsValue(nil)
		}
		return AsValue(v.getResolvedValue().Index(i).Interface())
	case reflect.String:
		return AsValue(v.getResolvedValue().Slice(i, i+1).Interface())
	default:
		logf("Value.Slice() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return AsValue([]int{})
	}
}

// Checks whether the underlying value (which must be of type struct, map,
// string, array or slice) contains of another Value (e. g. used to check
// whether a struct contains of a specific field or a map contains a specific key).
//
// Example:
//     AsValue("Hello, World!").Contains(AsValue("World")) == true
func (v *Value) Contains(other *Value) bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Struct:
		field_value := v.getResolvedValue().FieldByName(other.String())
		return field_value.IsValid()
	case reflect.Map:
		var map_value reflect.Value
		switch other.Interface().(type) {
		case int:
			map_value = v.getResolvedValue().MapIndex(other.getResolvedValue())
		case string:
			map_value = v.getResolvedValue().MapIndex(other.getResolvedValue())
		default:
			logf("Value.Contains() does not support lookup type '%s'\n", other.getResolvedValue().Kind().String())
			return false
		}

		return map_value.IsValid()
	case reflect.String:
		return strings.Contains(v.getResolvedValue().String(), other.String())

	// TODO: reflect.Array, reflect.Slice

	default:
		logf("Value.Contains() not available for type: %s\n", v.getResolvedValue().Kind().String())
		return false
	}
}

// Checks whether the underlying value is of type array, slice or string.
// You normally would use CanSlice() before using the Slice() operation.
func (v *Value) CanSlice() bool {
	switch v.getResolvedValue().Kind() {
	case reflect.Array, reflect.Slice, reflect.String:
		return true
	}
	return false
}

// Iterates over a map, array, slice or a string. It calls the
// function's first argument for every value with the following arguments:
//
//     idx      current 0-index
//     count    total amount of items
//     key      *Value for the key or item
//     value    *Value (only for maps, the respective value for a specific key)
//
// If the underlying value has no items or is not one of the types above,
// the empty function (function's second argument) will be called.
func (v *Value) Iterate(fn func(idx, count int, key, value *Value) bool, empty func()) {
	switch v.getResolvedValue().Kind() {
	case reflect.Map:
		keys := v.getResolvedValue().MapKeys()
		keyLen := len(keys)
		for idx, key := range keys {
			value := v.getResolvedValue().MapIndex(key)
			if !fn(idx, keyLen, &Value{key}, &Value{value}) {
				return
			}
		}
		if keyLen == 0 {
			empty()
		}
		return // done
	case reflect.Array, reflect.Slice:
		itemCount := v.getResolvedValue().Len()
		if itemCount > 0 {
			for i := 0; i < itemCount; i++ {
				if !fn(i, itemCount, &Value{v.getResolvedValue().Index(i)}, nil) {
					return
				}
			}
		} else {
			empty()
		}
		return // done
	case reflect.String:
		runeCount := v.getResolvedValue().Len()
		if runeCount > 0 {
			for i := 0; i < runeCount; i++ {
				if !fn(i, runeCount, &Value{v.getResolvedValue().Slice(i, i+1)}, nil) {
					return
				}
			}
		} else {
			empty()
		}
		return // done
	default:
		logf("Value.Iterate() not available for type: %s\n", v.getResolvedValue().Kind().String())
	}
	empty()
}

// Gives you access to the underlying value.
func (v *Value) Interface() interface{} {
	if v.v.IsValid() {
		return v.v.Interface()
	}
	return nil
}

// Checks whether two values are containing the same value or object.
func (v *Value) EqualValueTo(other *Value) bool {
	return v.Interface() == other.Interface()
}
