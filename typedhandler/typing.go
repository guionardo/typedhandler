package typedhandler

import "reflect"

// mustBeAPointer panics if the type T is not a pointer type.
func mustBeAPointer[T any]() {
	rInType := reflect.TypeFor[T]()
	if rInType.Kind() != reflect.Pointer {
		panic(rInType.Name() + " must be a pointer type")
	}
}

// typeName returns the fully qualified name of the type T,
// including package path.
func typeName[T any]() string {
	t := getType[T]()
	return t.PkgPath() + "." + t.Name()
}

// getType returns the underlying type of T
// If T is a pointer type, it returns the element type
func getType[T any]() reflect.Type {
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

// typeString returns a string representation of the type of v,
// including pointer or array/slice notation.
func typeString(v any) string {
	t := reflect.TypeOf(v)
	prefix := ""

	switch t.Kind() {
	case reflect.Pointer:
		prefix = "*"
		t = t.Elem()
	case reflect.Array, reflect.Slice:
		prefix = "[]"
		t = t.Elem()
	}

	return prefix + t.Name()
}

// areTypesCompatible checks if two types are compatible, ignoring pointer indirection.
// For example, *T and T are considered compatible.
func areTypesCompatible(t1, t2 reflect.Type) bool {
	if t1.Kind() == reflect.Pointer {
		t1 = t1.Elem()
	}

	if t2.Kind() == reflect.Pointer {
		t2 = t2.Elem()
	}

	return t1 == t2
}
