package typedhandler

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	Id        int
	Name      string
	Age       uint
	Distance  float64
	Enabled   bool
	CreatedAt time.Time
	TTL       time.Duration
	Id64      int64

	I8 int8
	U8 uint8

	I16 int16
	U16 uint16

	I32 int32
	U32 uint32

	I64 int64
	U64 uint64

	F32 float32
	F64 float64
}

func Test_convertData(t *testing.T) { //nolint
	t.Parallel()

	var (
		instance      TestStruct
		instanceValue = reflect.ValueOf(&instance).Elem()
	)

	tests := []struct {
		name       string
		data       string
		fieldIndex int
		check      func(t *testing.T)
	}{
		{
			name: "id:int", data: "1", fieldIndex: 0,
			check: func(t *testing.T) { assert.Equal(t, 1, instance.Id) },
		},
		{
			name: "id64:int64", data: "2", fieldIndex: 7,
			check: func(t *testing.T) {
				assert.Equal(t, int64(2), instanceValue.Field(7).Int())
			},
		},
		{
			name: "name:string", data: "John Doe", fieldIndex: 1,
			check: func(t *testing.T) {
				assert.Equal(t, "John Doe", instanceValue.Field(1).String())
			},
		},
		{
			name: "age:uint", data: "1", fieldIndex: 2,
			check: func(t *testing.T) {
				assert.Equal(t, uint64(1), instanceValue.Field(2).Uint())
			},
		},
		{
			name: "distance:float64", data: "100.5", fieldIndex: 3,
			check: func(t *testing.T) {
				assert.InEpsilon(t, 100.5, instanceValue.Field(3).Float(), 0.01)
			},
		},
		{
			name: "enabled:bool", data: "true", fieldIndex: 4,
			check: func(t *testing.T) {
				assert.True(t, instanceValue.Field(4).Bool())
			},
		},
		{
			name: "createdAt:time.Time", data: "2021-01-01T00:00:00Z", fieldIndex: 5,
			check: func(t *testing.T) {
				assert.Equal(
					t,
					time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
					instanceValue.Field(5).Interface().(time.Time),
				)
			},
		},
		{
			name: "ttl:time.Duration", data: "1h", fieldIndex: 6,
			check: func(t *testing.T) {
				assert.Equal(t, time.Hour, instanceValue.Field(6).Interface().(time.Duration))
			},
		},
		{
			name: "i8:int8", data: "127", fieldIndex: 8,
			check: func(t *testing.T) {
				assert.Equal(t, int8(127), int8(instanceValue.Field(8).Int())) //nolint
			},
		},
		{
			name: "u8:uint8", data: "255", fieldIndex: 9,
			check: func(t *testing.T) {
				assert.Equal(t, uint8(255), uint8(instanceValue.Field(9).Uint())) //nolint
			},
		},
		{
			name: "i16:int16", data: "32767", fieldIndex: 10,
			check: func(t *testing.T) {
				assert.Equal(t, int16(32767), int16(instanceValue.Field(10).Int())) //nolint
			},
		},
		{
			name: "u16:uint16", data: "65535", fieldIndex: 11,
			check: func(t *testing.T) {
				assert.Equal(t, uint16(65535), uint16(instanceValue.Field(11).Uint())) //nolint
			},
		},
		{
			name: "i32:int32", data: "2147483647", fieldIndex: 12,
			check: func(t *testing.T) {
				assert.Equal(t, int32(2147483647), int32(instanceValue.Field(12).Int())) //nolint
			},
		},
		{
			name: "u32:uint32", data: "4294967295", fieldIndex: 13,
			check: func(t *testing.T) {
				assert.Equal(t, uint32(4294967295), uint32(instanceValue.Field(13).Uint())) //nolint
			},
		},
		{
			name: "i64:int64", data: "9223372036854775807", fieldIndex: 14,
			check: func(t *testing.T) {
				assert.Equal(t, int64(9223372036854775807), instanceValue.Field(14).Int())
			},
		},
		{
			name: "u64:uint64", data: "18446744073709551615", fieldIndex: 15,
			check: func(t *testing.T) {
				assert.Equal(t, uint64(18446744073709551615), instanceValue.Field(15).Uint())
			},
		},
		{
			name: "f32:float32", data: "3.14", fieldIndex: 16,
			check: func(t *testing.T) {
				assert.InEpsilon(t, float32(3.14), float32(instanceValue.Field(16).Float()), 0.01)
			},
		},
		{
			name: "f64:float64", data: "2.718281828459045", fieldIndex: 17,
			check: func(t *testing.T) {
				assert.InEpsilon(t, 2.718281828459045, instanceValue.Field(17).Float(), 0.01)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			gotErr := convertData(test.data, test.fieldIndex, instanceValue)
			require.NoError(t, gotErr)
			test.check(t)
		})
	}
}
