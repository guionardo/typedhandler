package typedhandler

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

const (
	bit64 = 64
	bit32 = 32
	bit16 = 16
	bit8  = 8
)

var (
	timeType     = reflect.TypeFor[time.Time]()
	durationType = reflect.TypeFor[time.Duration]()
	TimeFormats  = []string{time.RFC3339, time.RFC3339Nano, time.RFC1123, time.RFC1123Z, time.ANSIC}
)

// convertData converts the data to the appropriate type and sets it in the struct
// A query value can be: string, int, uint, float64, bool, time.Time, time.Duration
func convertData(data string, fieldIndex int, structValue reflect.Value) (err error) {
	fieldType := structValue.Field(int(fieldIndex)).Type()
	field := structValue.Field(int(fieldIndex))

	// Handle special types first
	switch fieldType {
	case timeType:
		return convertTime(data, field)
	case durationType:
		return convertDuration(data, field)
	}

	// Handle primitive types by kind
	return convertByKind(data, field)
}

// convertByKind handles conversion based on the reflect.Kind
func convertByKind(data string, field reflect.Value) error {
	kind := field.Kind()

	// Group similar types together
	switch {
	case kind == reflect.String:
		field.SetString(data)
		return nil

	case kind == reflect.Bool:
		return convertBool(data, field)

	case isIntKind(kind):
		return convertInt(data, getBitSize(kind), field)

	case isUintKind(kind):
		return convertUint(data, getBitSize(kind), field)

	case isFloatKind(kind):
		return convertFloat(data, getBitSize(kind), field)

	default:
		return fmt.Errorf("unsupported field type: %s", field.Type().Name())
	}
}

// isIntKind checks if the kind is an integer type
func isIntKind(kind reflect.Kind) bool {
	return kind == reflect.Int || kind == reflect.Int8 ||
		kind == reflect.Int16 || kind == reflect.Int32 ||
		kind == reflect.Int64
}

// isUintKind checks if the kind is an unsigned integer type
func isUintKind(kind reflect.Kind) bool {
	return kind == reflect.Uint || kind == reflect.Uint8 ||
		kind == reflect.Uint16 || kind == reflect.Uint32 ||
		kind == reflect.Uint64
}

// isFloatKind checks if the kind is a float type
func isFloatKind(kind reflect.Kind) bool {
	return kind == reflect.Float32 || kind == reflect.Float64
}

// getBitSize returns the bit size for numeric types
func getBitSize(kind reflect.Kind) int {
	switch kind {
	case reflect.Int8, reflect.Uint8:
		return bit8
	case reflect.Int16, reflect.Uint16:
		return bit16
	case reflect.Int32, reflect.Uint32, reflect.Float32:
		return bit32
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64, reflect.Float64:
		return bit64
	default:
		return bit64 // default fallback
	}
}

// convertBool converts a string to a boolean and sets it in the struct field
func convertBool(data string, field reflect.Value) (err error) {
	var boolValue bool
	if boolValue, err = strconv.ParseBool(data); err == nil {
		field.SetBool(boolValue)
	}

	return err
}

// convertDuration converts a string to a time.Duration and sets it in the struct field
func convertDuration(data string, field reflect.Value) (err error) {
	var durationValue time.Duration
	if durationValue, err = time.ParseDuration(data); err == nil {
		field.SetInt(int64(durationValue))
	}

	return err
}

func convertInt(data string, bitSize int, field reflect.Value) (err error) {
	var intValue int64
	if intValue, err = strconv.ParseInt(data, 10, bitSize); err == nil {
		field.SetInt(intValue)
	}

	return err
}

func convertUint(data string, bitSize int, field reflect.Value) (err error) {
	var uintValue uint64
	if uintValue, err = strconv.ParseUint(data, 10, bitSize); err == nil {
		field.SetUint(uintValue)
	}

	return err
}

func convertFloat(data string, bitSize int, field reflect.Value) (err error) {
	var floatValue float64
	if floatValue, err = strconv.ParseFloat(data, bitSize); err == nil {
		field.SetFloat(floatValue)
	}

	return err
}

func convertTime(data string, field reflect.Value) (err error) {
	var timeValue time.Time
	if timeValue, err = ParseTime(data); err == nil {
		field.Set(reflect.ValueOf(timeValue))
	}

	return err
}
