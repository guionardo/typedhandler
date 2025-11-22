package typedhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

type (
	ParserInfos struct {
		dataType     reflect.Type   // type of the request schema
		hasValidate  bool           // whether the request schema has validate tag
		queryFields  map[int]string // query fields
		pathFields   map[int]string // path fields
		headerFields map[int]string // header fields

		bodyType BodyType

		parseBodyFunc func(r *http.Request, instance any) error

		errors error
	}

	BodyType uint8
)

const (
	NoBody    BodyType = iota // No body from request
	JsonBody                  // Body is unmarhaled into hole struct
	JsonField                 // Body is unmarshaled into struct field referenced by BodyFieldGetter interface
)

var (
	infos = make(map[string]*ParserInfos)
	mu    sync.Mutex
)

func getParserInfos[RIn RequestSchema]() *ParserInfos {
	mu.Lock()
	defer mu.Unlock()

	t := typeName[RIn]()
	if pi, ok := infos[t]; ok {
		return pi
	}

	pi := createParserInfos[RIn]()
	infos[t] = pi

	return pi
}

func (pi *ParserInfos) checkValidate(field reflect.StructField) *ParserInfos {
	if !pi.hasValidate && len(field.Tag.Get("validate")) > 0 {
		pi.hasValidate = true
	}

	return pi
}

func (pi *ParserInfos) checkQuery(field reflect.StructField) *ParserInfos {
	if formField := field.Tag.Get("form"); len(formField) > 0 {
		pi.queryFields[field.Index[0]] = formField
	}

	if queryField := field.Tag.Get("query"); len(queryField) > 0 {
		pi.queryFields[field.Index[0]] = queryField
	}

	return pi
}

func (pi *ParserInfos) checkPath(field reflect.StructField) *ParserInfos {
	if pathParam := field.Tag.Get("path"); len(pathParam) > 0 {
		pi.pathFields[field.Index[0]] = pathParam
	}

	return pi
}

func (pi *ParserInfos) checkHeader(field reflect.StructField) *ParserInfos {
	if headerField := field.Tag.Get("header"); len(headerField) > 0 {
		if field.Type.Kind() != reflect.String {
			pi.errors = errors.Join(pi.errors, fmt.Errorf("header field %s must be a string", field.Name))
		} else {
			pi.headerFields[field.Index[0]] = headerField
		}
	}

	return pi
}

func (pi *ParserInfos) checkJson(field reflect.StructField) *ParserInfos {
	if jsonBody := field.Tag.Get("json"); len(jsonBody) > 0 {
		// a json tag implies that the request body will be parsed into hole instance
		pi.bodyType = JsonBody
		pi.parseBodyFunc = parseBodyInstance
	}

	return pi
}

func (pi *ParserInfos) checkBody(field reflect.StructField, instance any) *ParserInfos {
	if bodyField := field.Tag.Get("body"); len(bodyField) == 0 {
		return pi
	}

	// a body tag implies that the request body will be unmarshaled into a struct field
	// Check if the field implements the BodyFieldGetter interface
	bfg, ok := any(instance).(BodyFieldGetter)
	if !ok {
		pi.errors = errors.Join(
			pi.errors,
			fmt.Errorf(
				"type %s.%s: must implement the BodyFieldGetter interface returning a pointer to field %s",
				pi.dataType.PkgPath(),
				pi.dataType.Name(),
				field.Name,
			),
		)

		return pi
	}

	bodyFieldInstance := bfg.GetBodyField()

	t := reflect.TypeOf(bodyFieldInstance)
	if !areTypesCompatible(field.Type, t) {
		pi.errors = errors.Join(pi.errors,
			fmt.Errorf("%s.GetBodyField() is returning a type %s incompatible to field %s %s",
				typeString(instance), typeString(bodyFieldInstance), field.Name, field.Type.Name()))

		return pi
	}

	pi.parseBodyFunc = parseBodyField
	pi.bodyType = JsonField

	return pi
}

func (pi *ParserInfos) checkParseableFields(instance any) {
	value := reflect.ValueOf(instance)
	visited := make(map[int]struct{})
	checkPF := func(f map[int]string) {
		for fieldIndex := range f {
			if _, ok := visited[fieldIndex]; !ok && !value.Elem().Field(fieldIndex).CanSet() {
				pi.errors = errors.Join(
					pi.errors,
					fmt.Errorf("field %s is not settable", pi.dataType.Field(fieldIndex).Name),
				)
			}
		}
	}
	checkPF(pi.queryFields)
	checkPF(pi.pathFields)
	checkPF(pi.headerFields)
}

func createParserInfos[RIn RequestSchema]() *ParserInfos {
	t := getType[RIn]()
	parserInfos := &ParserInfos{
		dataType:     t,
		queryFields:  make(map[int]string),
		pathFields:   make(map[int]string),
		headerFields: make(map[int]string),
	}

	instance := NewInstance[RIn](t)
	for i := range t.NumField() {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		parserInfos.
			checkValidate(field).
			checkQuery(field).
			checkPath(field).
			checkHeader(field).
			checkJson(field).
			checkBody(field, instance)
	}

	parserInfos.checkParseableFields(instance)

	if parserInfos.errors != nil {
		panic(fmt.Errorf("%s.%s: %w", t.PkgPath(), t.Name(), parserInfos.errors))
	}

	return parserInfos
	// return &ParserInfos{
	// 	dataType:       t,
	// 	hasValidate:    hasValidate,
	// 	queryFields:    queryFields,
	// 	pathFields:     pathFields,
	// 	bodyFieldIndex: bodyFieldIndex,
	// 	isJson:         isJson,
	// 	hasBody:        fieldCount > 0,
	// 	headerFields:   headerFields,
	// 	parseBodyFunc:  parseBodyFunc,
	// }
}

func typeName[T any]() string {
	t := getType[T]()
	return t.PkgPath() + "/" + t.Name()
}

func getType[T any]() reflect.Type {
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

func typeString(v any) string {
	t := reflect.TypeOf(v)
	prefix := ""

	switch t.Kind() {
	case reflect.Pointer:
		prefix = "*"
		t = t.Elem()
	case reflect.Array:
		prefix = "[]"
		t = t.Elem()
	}

	return prefix + t.Name()
}

func parseBodyField(r *http.Request, instance any) error {
	var rawInstance = instance

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

func parseBodyInstance(r *http.Request, instance any) error {
	return json.NewDecoder(r.Body).Decode(instance)
}

func areTypesCompatible(t1, t2 reflect.Type) bool {
	if t1.Kind() == reflect.Pointer {
		t1 = t1.Elem()
	}

	if t2.Kind() == reflect.Pointer {
		t2 = t2.Elem()
	}

	return t1 == t2
}
