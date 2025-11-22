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

func CreateParser[RIn RequestSchema]() ParseRequestFunc[RIn] {
	infos := getParserInfos[RIn]()

	resetFunc := createResetFunc[RIn](infos)
	pool := NewInstancePool(resetFunc)

	return func(r *http.Request) (instance RIn, err error) {
		instance = NewInstance[RIn](infos.dataType)

		ptrValue := reflect.ValueOf(instance)
		structValue := ptrValue.Elem()

		err = parseBody(r, infos, instance)
		if err == nil {
			err = parseHeaders(r, infos, structValue)
		}

		if err == nil {
			err = parseQuery(r, infos, structValue)
		}

		if err == nil {
			err = parsePath(r, infos, structValue)
		}

		pool.Put(instance)

		return instance, err
	}
}

func parseBody[RIn RequestSchema](r *http.Request, infos *ParserInfos, instance RIn) error {
	if infos.bodyType != NoBody {
		return infos.parseBodyFunc(r, instance)
	}

	return nil
}

// parseHeaders parses the headers and sets the values in the struct
// A header value is always a string
func parseHeaders(r *http.Request, infos *ParserInfos, structValue reflect.Value) error {
	if len(infos.headerFields) == 0 {
		return nil
	}

	for key, value := range infos.headerFields {
		header := r.Header.Get(value)
		if err := convertData(header, int(key), structValue); err != nil {
			return err
		}
	}

	return nil
}

// parseQuery parses the query and sets the values in the struct
// A query value can be: string, int, uint, float64, bool, time.Time, time.Duration
func parseQuery(r *http.Request, infos *ParserInfos, structValue reflect.Value) error {
	if len(infos.queryFields) == 0 {
		return nil
	}

	for key, value := range infos.queryFields {
		query := r.URL.Query().Get(value)
		if err := convertData(query, int(key), structValue); err != nil {
			return err
		}
	}

	return nil
}

// parsePath parses the path and sets the values in the struct
// A path value can be: string, int, uint, float64, bool, time.Time, time.Duration
func parsePath(r *http.Request, infos *ParserInfos, structValue reflect.Value) error {
	if len(infos.pathFields) == 0 {
		return nil
	}

	for key, value := range infos.pathFields {
		path := r.PathValue(value)
		if err := convertData(path, int(key), structValue); err != nil {
			return err
		}
	}

	return nil
}
