package typedhandler

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	typeNameSample  struct{}
	typeNameSample2 struct{}
)

func TestMustBeAPointer(t *testing.T) {
	t.Parallel()
	t.Run("must_be_a_pointer_panics_for_non_pointer", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			mustBeAPointer[typeNameSample]()
		})
	})
	t.Run("must_be_a_pointer_does_not_panic_for_pointer", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			mustBeAPointer[*typeNameSample]()
		})
	})
}

func Test_typeName(t *testing.T) {
	t.Parallel()

	got := typeName[typeNameSample]()
	assert.Equal(t, "github.com/guionardo/typedhandler/typedhandler.typeNameSample", got)
}

func Test_getType(t *testing.T) {
	t.Parallel()
	t.Run("get_type_returns_element_type_for_pointer", func(t *testing.T) {
		t.Parallel()

		got := getType[*typeNameSample]()
		assert.Equal(t, reflect.TypeFor[typeNameSample](), got)
	})
	t.Run("get_type_returns_type_for_non_pointer", func(t *testing.T) {
		t.Parallel()

		got := getType[typeNameSample]()
		assert.Equal(t, reflect.TypeFor[typeNameSample](), got)
	})
}

func Test_typeString(t *testing.T) {
	t.Parallel()
	t.Run("concrete_type", func(t *testing.T) {
		t.Parallel()

		got := typeString(typeNameSample{})
		assert.Equal(t, "typeNameSample", got)
	})
	t.Run("pointer_type", func(t *testing.T) {
		t.Parallel()

		got := typeString(&typeNameSample{})
		assert.Equal(t, "*typeNameSample", got)
	})
	t.Run("array_type", func(t *testing.T) {
		t.Parallel()

		got := typeString([]typeNameSample{})
		assert.Equal(t, "[]typeNameSample", got)
	})
}

func Test_areTypesCompatible(t *testing.T) { //nolint
	t.Parallel()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		t1   reflect.Type
		t2   reflect.Type
		want bool
	}{
		{
			name: "same_concrete_types",
			t1:   reflect.TypeFor[typeNameSample](),
			t2:   reflect.TypeFor[typeNameSample](),
			want: true,
		},
		{
			name: "same_pointer_types",
			t1:   reflect.TypeFor[*typeNameSample](),
			t2:   reflect.TypeFor[*typeNameSample](),
			want: true,
		},
		{
			name: "concrete_and_pointer_types",
			t1:   reflect.TypeFor[typeNameSample](),
			t2:   reflect.TypeFor[*typeNameSample](),
			want: true,
		},
		{
			name: "different_concrete_types",
			t1:   reflect.TypeFor[typeNameSample](),
			t2:   reflect.TypeFor[typeNameSample2](),
			want: false,
		},
		{
			name: "different_pointer_types",
			t1:   reflect.TypeFor[*typeNameSample](),
			t2:   reflect.TypeFor[*typeNameSample2](),
			want: false,
		},
	}
	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			t.Parallel()

			got := areTypesCompatible(tests[i].t1, tests[i].t2)
			assert.Equal(t, tests[i].want, got, "areTypesCompatible(%v, %v)", tests[i].t1, tests[i].t2)
		})
	}
}
