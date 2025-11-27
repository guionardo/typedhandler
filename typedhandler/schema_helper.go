package typedhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/go-playground/validator/v10"
)

// SchemaHelper is a helper for request schema RIn
type (
	SchemaHelper[RIn RequestSchema] struct {
		queryFields  map[int]string // query fields
		pathFields   map[int]string // path fields
		headerFields map[int]string // header fields

		typeFor       reflect.Type
		bodyType      BodyType
		ResetFunc     func(RIn)
		parseBodyFunc func(r *http.Request, instance any) error
		validateFunc  func(RIn) error

		instancePool sync.Pool
		poolGetFunc  func() any
		poolPutFunc  func(any)

		hasValidate bool
		errors      error

		instanceCount atomic.Uint64
	}

	BodyType uint8
)

const (
	NoBody    BodyType = iota // No body from request
	JsonBody                  // Body is unmarhaled into hole struct
	JsonField                 // Body is unmarshaled into struct field referenced by BodyFieldGetter interface
)

var (
	schemaHelpers map[string]any = make(map[string]any)
	shMu          sync.Mutex
	// PoolEnabled for reusable instances - normally enabled
	PoolEnabled = true
)

// GetSchemaHelper returns the SchemaHelper for request schema RIn
// It creates a new SchemaHelper if it does not exist
// RIn must be a pointer type
func GetSchemaHelper[RIn RequestSchema]() *SchemaHelper[RIn] {
	mustBeAPointer[RIn]()

	shMu.Lock()
	defer shMu.Unlock()

	t := typeName[RIn]()

	instance, found := schemaHelpers[t]
	if found {
		return instance.(*SchemaHelper[RIn])
	}

	fieldCount := getType[RIn]().NumField()
	// Create new SchemaHelper
	helper := &SchemaHelper[RIn]{
		queryFields:  make(map[int]string, fieldCount),
		pathFields:   make(map[int]string, fieldCount),
		headerFields: make(map[int]string, fieldCount),
		typeFor:      reflect.TypeFor[RIn](),
	}
	helper.initializeFields()

	if helper.errors != nil {
		t := getType[RIn]()
		err := fmt.Errorf("%s.%s: %w", t.PkgPath(), t.Name(), helper.errors)
		slog.Error("GetSchemaHelper", slog.String("type", t.PkgPath()+"."+t.Name()), slog.Any("error", helper.errors))
		panic(err)
	}

	helper.createResetFunc()
	helper.createValidateFunc()
	helper.createInstancePool()

	schemaHelpers[t] = helper

	return helper
}

// GetInstance gets an instance of RIn from the pool or creates a new one
func (sh *SchemaHelper[RIn]) GetInstance() RIn {
	return sh.poolGetFunc().(RIn)
}

// PutInstance returns an instance of RIn to the pool
func (sh *SchemaHelper[RIn]) PutInstance(instance RIn) {
	sh.ResetFunc(instance)
	sh.poolPutFunc(instance)
}

// newInstance creates a new instance of RIn
func (sh *SchemaHelper[RIn]) newInstance() (instance RIn) {
	// Get the underlying struct type (strip pointer if dataType is a pointer)
	structType := sh.typeFor
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	// RIn is a pointer type (validated by MustBeAPointer),
	// create a pointer to the struct
	// reflect.New(structType) returns a pointer to structType
	ptr := reflect.New(structType)

	return ptr.Interface().(RIn)
}

// Errors returns any errors found during SchemaHelper initialization
func (sh *SchemaHelper[RIn]) Errors() error {
	return sh.errors
}

// initializeFields initializes the SchemaHelper fields by inspecting the struct tags
func (sh *SchemaHelper[RIn]) initializeFields() {
	t := getType[RIn]()
	instance := sh.newInstance()

	for i := range t.NumField() {
		field := t.Field(i)
		if field.IsExported() {
			sh.checkValidate(&field).
				checkQuery(&field).
				checkPath(&field).
				checkHeader(&field).
				checkJson(&field).
				checkBody(&field, instance)
		}
	}

	sh.checkParseableFields(instance)
}

// checkValidate identifies if the struct has any validate tags
func (sh *SchemaHelper[RIn]) checkValidate(field *reflect.StructField) *SchemaHelper[RIn] {
	if !sh.hasValidate && field.Tag.Get("validate") != "" {
		sh.hasValidate = true
	}

	return sh
}

// checkQuery identifies query fields from struct tags "form" and "query"
func (sh *SchemaHelper[RIn]) checkQuery(field *reflect.StructField) *SchemaHelper[RIn] {
	if formField := field.Tag.Get("form"); formField != "" {
		sh.queryFields[field.Index[0]] = formField
	}

	if queryField := field.Tag.Get("query"); queryField != "" {
		sh.queryFields[field.Index[0]] = queryField
	}

	return sh
}

// checkPath identifies path fields from struct tags "path"
func (sh *SchemaHelper[RIn]) checkPath(field *reflect.StructField) *SchemaHelper[RIn] {
	if pathParam := field.Tag.Get("path"); pathParam != "" {
		sh.pathFields[field.Index[0]] = pathParam
	}

	return sh
}

func (sh *SchemaHelper[RIn]) checkHeader(field *reflect.StructField) *SchemaHelper[RIn] {
	if headerField := field.Tag.Get("header"); headerField != "" {
		if field.Type.Kind() != reflect.String {
			sh.errors = errors.Join(sh.errors, fmt.Errorf("header field %s must be a string", field.Name))
		} else {
			sh.headerFields[field.Index[0]] = headerField
		}
	}

	return sh
}

func (sh *SchemaHelper[RIn]) checkJson(field *reflect.StructField) *SchemaHelper[RIn] {
	if jsonBody := field.Tag.Get("json"); jsonBody != "" {
		// a json tag implies that the request body will be parsed into hole instance
		sh.bodyType = JsonBody
		sh.parseBodyFunc = parseBodyInstance
	}

	return sh
}

func (sh *SchemaHelper[RIn]) checkBody(field *reflect.StructField, instance any) *SchemaHelper[RIn] {
	if bodyField := field.Tag.Get("body"); bodyField == "" {
		return sh
	}

	// a body tag implies that the request body will be unmarshaled into a struct field
	// Check if the field implements the BodyFieldGetter interface
	bfg, ok := any(instance).(BodyFieldGetter)
	if !ok {
		sh.errors = errors.Join(
			sh.errors,
			fmt.Errorf(
				"type %s.%s: must implement the BodyFieldGetter interface returning a pointer to field %s",
				sh.typeFor.PkgPath(),
				sh.typeFor.Name(),
				field.Name,
			),
		)

		return sh
	}

	bodyFieldInstance := bfg.GetBodyField()

	t := reflect.TypeOf(bodyFieldInstance)
	if !areTypesCompatible(field.Type, t) {
		sh.errors = errors.Join(sh.errors,
			fmt.Errorf("%s.GetBodyField() is returning a type %s incompatible to field %s %s",
				typeString(instance), typeString(bodyFieldInstance), field.Name, field.Type.Name()))

		return sh
	}

	sh.parseBodyFunc = parseBodyField
	sh.bodyType = JsonField

	return sh
}

func (sh *SchemaHelper[RIn]) checkParseableFields(instance any) {
	value := reflect.ValueOf(instance)
	t := reflect.TypeOf(instance)
	visited := make(map[int]struct{})
	checkPF := func(f map[int]string) {
		for fieldIndex := range f {
			if _, ok := visited[fieldIndex]; ok {
				continue
			}

			field := value.Elem().Field(fieldIndex)
			slog.Debug("checkParseableFields", slog.String("type", t.String()), slog.String("field", field.String()))

			if !field.CanSet() { //!value.Elem().Field(fieldIndex).CanSet() {
				sh.errors = errors.Join(
					sh.errors,
					fmt.Errorf("field %s is not settable", sh.typeFor.Field(fieldIndex).Name),
				)
			}

			visited[fieldIndex] = struct{}{}
		}
	}
	checkPF(sh.queryFields)
	checkPF(sh.pathFields)
	checkPF(sh.headerFields)
}

func (sh *SchemaHelper[RIn]) createResetFunc() {
	if (len(sh.headerFields) + len(sh.queryFields) + len(sh.pathFields)) == 0 {
		// Only create reset function if we have non-body fields that need clearing
		sh.ResetFunc = func(RIn) {} // NOOP
		return
	}

	structType := getType[RIn]() // sh.typeFor

	var zero RIn

	// Check if type implements Resettable interface (compile-time check)
	if _, ok := any(zero).(Resettable); ok {
		// Use the Resettable interface method - no reflection needed
		sh.ResetFunc = func(instance RIn) {
			any(instance).(Resettable).Reset()
		}

		return
	}

	rInType := reflect.TypeOf(zero)

	if rInType.Kind() == reflect.Pointer {
		// For pointer types, reset all fields
		sh.ResetFunc = func(instance RIn) {
			ptrValue := reflect.ValueOf(instance)
			if ptrValue.IsNil() {
				return
			}

			structValue := ptrValue.Elem()
			zeroValue := reflect.Zero(structType)
			structValue.Set(zeroValue)
		}

		return
	}
	// For value types, reset all fields
	panic("value type RIn in SchemaHelper createResetFunc")
}

func (sh *SchemaHelper[RIn]) createValidateFunc() {
	if !sh.hasValidate {
		// No validate tags found, no validate function needed
		sh.validateFunc = func(RIn) error { return nil } // NOOP
		return
	}

	var zero RIn
	if validatable, ok := any(zero).(Validatable); ok {
		sh.validateFunc = func(instance RIn) error {
			return validatable.Validate()
		}

		return
	}

	v := validator.New(validator.WithRequiredStructEnabled())

	sh.validateFunc = func(instance RIn) error {
		return v.Struct(instance)
	}
}

func (sh *SchemaHelper[RIn]) createInstancePool() {
	if PoolEnabled {
		sh.instancePool = sync.Pool{
			New: func() any {
				sh.instanceCount.Add(1)
				return sh.newInstance()
			},
		}
		sh.poolGetFunc = sh.instancePool.Get
		sh.poolPutFunc = sh.instancePool.Put
	} else {
		sh.poolGetFunc = func() any {
			sh.instanceCount.Add(1)
			return sh.newInstance()
		}
		sh.poolPutFunc = func(any) {
			sh.instanceCount.Store(sh.instanceCount.Load() - 1)
		}
	}
}

// parseRequestBody parses the body from request and sets the values in the struct
// The body can be JSON unmarshaled into the whole struct or into a struct field
func (sh *SchemaHelper[RIn]) parseRequestBody(r *http.Request, instance RIn) error {
	if sh.bodyType != NoBody {
		return sh.parseBodyFunc(r, instance)
	}

	return nil
}

// parseRequestHeaders parses the headers and sets the values in the struct
// A header value is always a string
func (sh *SchemaHelper[RIn]) parseRequestHeaders(r *http.Request, structValue reflect.Value) (err error) {
	var header string
	for key, value := range sh.headerFields {
		header = r.Header.Get(value)
		if err = convertData(header, int(key), structValue); err != nil {
			break
		}
	}

	return err
}

// parseRequestPath parses the path and sets the values in the struct
// A path value can be: string, int, uint, float64, bool, time.Time, time.Duration
func (sh *SchemaHelper[RIn]) parseRequestPath(r *http.Request, structValue reflect.Value) (err error) {
	var path string
	for key, value := range sh.pathFields {
		path = r.PathValue(value)
		if err = convertData(path, int(key), structValue); err != nil {
			break
		}
	}

	return err
}

// parseRequestQuery parses the query and sets the values in the struct
// A query value can be: string, int, uint, float64, bool, time.Time, time.Duration
func (sh *SchemaHelper[RIn]) parseRequestQuery(r *http.Request, structValue reflect.Value) (err error) {
	var query string
	for key, value := range sh.queryFields {
		query = r.URL.Query().Get(value)
		if err = convertData(query, int(key), structValue); err != nil {
			break
		}
	}

	return err
}

// parseBodyInstance parses the body from request into the instance
func parseBodyInstance(r *http.Request, instance any) error {
	return json.NewDecoder(r.Body).Decode(instance)
}

// parseBodyField parses the body from request into the struct field returned by GetBodyField function
// the instance must implement the BodyFieldGetter interface
func parseBodyField(r *http.Request, instance any) error {
	rawInstance := instance

	bfg, ok := rawInstance.(BodyFieldGetter)
	if !ok {
		return fmt.Errorf("request schema %s has a body field, but does not implements the %s interface",
			typeString(instance),
			reflect.TypeFor[BodyFieldGetter]().Name())
	}

	bodyFieldValue := bfg.GetBodyField()
	if bodyFieldValue == nil {
		return fmt.Errorf("request schema %s must return the body field value as a pointer",
			typeString(instance))
	}

	return parseBodyInstance(r, bodyFieldValue)
}
