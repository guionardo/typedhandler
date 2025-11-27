package typedhandler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type (
	HandlerFunc                                         func(w http.ResponseWriter, r *http.Request)
	ServiceFunc[RIn RequestSchema, ROut ResponseSchema] func(ctx context.Context, request RIn) (response ROut, status int, err error) //nolint
	ParseRequestFunc[RIn RequestSchema]                 func(r *http.Request) (instance RIn, err error)
)

// CreateHandler creates a typed HTTP handler with the provided request parser, done function, and service function.
// The request parser is responsible for parsing the incoming HTTP request into the specified request schema type RIn.
// The done function is called at the end of the request handling, if provided.
// The service function processes the parsed request and returns a response of type ROut, along with an HTTP status code
// and an error if any.
//
// Where it's used:
//   - Examples: See `examples/simple/main.go` for a runnable example that calls
//     requestParser, doneFunc := typedhandler.CreateParser[*LoginRequest]()
//     handler := typedhandler.CreateHandler(requestParser, doneFunc, serviceFunc)
//   - Quick start: README.md includes a short usage example that demonstrates
//     creating a parser and passing it to `CreateHandler`.
//   - Tests & Benchmarks: The behavior of `CreateHandler` is exercised in
//     `typedhandler/handler_test.go` and `typedhandler/handler_create_handler_test.go`
//     (unit tests and benchmarks).
//   - Convenience wrapper: `CreateSimpleHandler` calls `CreateParser` and then
//     delegates to `CreateHandler`, so you can use `CreateSimpleHandler` when
//     you prefer the library to build the parser for you.
//
// In short, `CreateHandler` is the core function that composes parsing,
// business logic (service function), and response/error writing into a
// standard `http.HandlerFunc` usable with `http.HandleFunc` or any
// net/http-compatible router.
func CreateHandler[RIn RequestSchema, ROut ResponseSchema](
	parseRequestFunc ParseRequestFunc[RIn], doneFunc func(),
	serviceFunc ServiceFunc[RIn, ROut],
) HandlerFunc {
	mustBeAPointer[RIn]()

	var (
		zero     RIn
		preParse func(*http.Request) error
	)
	if preParseable, ok := any(zero).(PreParseable); ok {
		preParse = preParseable.PreParse
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if doneFunc != nil {
			defer doneFunc()
		}
		// pre-parse the request
		if preParse != nil {
			if err := preParse(r); err != nil {
				writeErrorResponse(w, err)
				return
			}
		}
		instance, err := parseRequestFunc(r)
		if err != nil {
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
	parserFunc, doneFunc := CreateParser[RIn]()
	return CreateHandler(parserFunc, doneFunc, serviceFunc)
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
		jsonError     HttpJsonError
		httpError     HttpError
		validateError validator.ValidationErrors
	)
	switch {
	case errors.As(err, &validateError):
		validationErrorToHttpJsonError(err, w)
	case errors.As(err, &jsonError):
		w.Header().Set("Content-Type", "application/json")
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

func validationErrorToHttpJsonError(err error, w http.ResponseWriter) {
	var validateError validator.ValidationErrors
	if errors.As(err, &validateError) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		sb := strings.Builder{}
		sb.WriteString("[")

		for i := range validateError {
			before, after, found := strings.Cut(validateError[i].Error(), " Error:")
			if found {
				sb.WriteString("\"" + after + "\"")
			} else {
				sb.WriteString("\"" + before + "\"")
			}

			if i < len(validateError)-1 {
				sb.WriteString(",")
			}
		}

		sb.WriteString("]")
		_, _ = w.Write([]byte(sb.String()))
	}
}
