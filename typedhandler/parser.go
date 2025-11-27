package typedhandler

import (
	"net/http"
	"reflect"
)

type (
	// BodyFieldGetter represents a struct that can get a pointer to a body field
	// This pointer will be used to unmarshal the body into the field
	BodyFieldGetter interface {
		GetBodyField() any
	}
)

// CreateParser creates a ParseRequestFunc and a doneFunc for request schema RIn
// RIn must be a pointer type
// The doneFunc must be called when the parser is no longer needed to release resources
func CreateParser[RIn RequestSchema]() (parserFunc ParseRequestFunc[RIn], doneFunc func()) {
	schemaHelper := GetSchemaHelper[RIn]()

	instance := schemaHelper.GetInstance()

	return func(r *http.Request) (RIn, error) {
			ptrValue := reflect.ValueOf(instance)
			structValue := ptrValue.Elem()

			var err error
			if err = schemaHelper.parseRequestBody(r, instance); err == nil {
				err = schemaHelper.parseRequestHeaders(r, structValue)
			}

			if err == nil {
				err = schemaHelper.parseRequestPath(r, structValue)
			}

			if err == nil {
				err = schemaHelper.parseRequestQuery(r, structValue)
			}

			if err == nil {
				err = schemaHelper.validateFunc(instance)
			}

			return instance, err
		}, func() {
			schemaHelper.PutInstance(instance)
		}
}
