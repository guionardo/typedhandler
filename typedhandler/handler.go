package typedhandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
)

type (
	HandlerFunc                                         func(w http.ResponseWriter, r *http.Request)
	ServiceFunc[RIn RequestSchema, ROut ResponseSchema] func(ctx context.Context, request RIn) (response ROut, status int, err error) //nolint
	ParseRequestFunc[RIn RequestSchema]                 func(r *http.Request) (instance RIn, err error)
	RequestSchema                                       any
	ResponseSchema                                      any
	HttpError                                           interface {
		error
		Status() int
	}
	HttpJsonError interface {
		HttpError
		Json() []byte
	}
)

func CreateHandler[RIn RequestSchema, ROut ResponseSchema](
	parseRequestFunc ParseRequestFunc[RIn],
	serviceFunc ServiceFunc[RIn, ROut],
) HandlerFunc {
	t := reflect.TypeFor[RIn]()
	if t.Kind() != reflect.Pointer {
		panic("CreateHandler must receive *" + t.Elem().Name())
	}

	resetFunc := createResetFunc[RIn](getParserInfos[RIn]())
	pool := NewInstancePool(resetFunc)

	return func(w http.ResponseWriter, r *http.Request) {
		instance, err := parseRequestFunc(r)
		if err != nil {
			pool.Put(instance)

			writeErrorResponse(w, err)

			return
		}

		response, status, err := serviceFunc(r.Context(), instance)
		if err == nil {
			err = writeResponse(w, status, response)
		}

		writeErrorResponse(w, err)
	}
}

func CreateSimpleHandler[RIn RequestSchema, ROut ResponseSchema](serviceFunc ServiceFunc[RIn, ROut]) HandlerFunc {
	parserFunc := CreateParser[RIn]()
	return CreateHandler(parserFunc, serviceFunc)
}

func writeResponse[ROut ResponseSchema](w http.ResponseWriter, status int, response ROut) error {
	if status <= 0 {
		status = http.StatusOK
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return err
	}

	w.WriteHeader(status)
	_, err = w.Write(responseBody)

	return err
}

func writeErrorResponse(w http.ResponseWriter, err error) {
	var (
		jsonError HttpJsonError
		httpError HttpError
	)
	switch {
	case errors.As(err, &jsonError):
		w.WriteHeader(jsonError.Status())
		_, _ = w.Write(jsonError.Json())
	case errors.As(err, &httpError):
		w.WriteHeader(httpError.Status())
		_, _ = w.Write([]byte(httpError.Error()))
	default:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	}
}
