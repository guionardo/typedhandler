package typedhandler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/guionardo/typedhandler/examples/sample"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	httpError struct {
		StatusCode int
		Message    string
	}
	jsonError struct {
		httpError
	}
)

func (e httpError) Status() int {
	return e.StatusCode
}

func (e httpError) Error() string {
	return e.Message
}

func (e jsonError) Json() []byte {
	return fmt.Appendf(nil, "{\"message\":\"%s\"}", e.Message)
}

func Test_writeErrorResponse(t *testing.T) {
	t.Parallel()
	t.Run("no_error", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		writeErrorResponse(w, nil)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
	t.Run("error", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		writeErrorResponse(w, errors.New("Bad Request"))
		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
		assert.Equal(t, "Bad Request", w.Body.String())
	})
	t.Run("http_error", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		writeErrorResponse(w, httpError{StatusCode: http.StatusBadRequest, Message: "Bad Request"})
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Equal(t, "Bad Request", w.Body.String())
	})
	t.Run("http_json_error", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		writeErrorResponse(
			w,
			jsonError{httpError: httpError{StatusCode: http.StatusBadRequest, Message: "Bad Request"}},
		)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.JSONEq(t, "{\"message\":\"Bad Request\"}", w.Body.String())
	})
	t.Run("validation_error", func(t *testing.T) {
		t.Parallel()

		type validationError struct {
			Name string `validate:"required"`
			Age  int    `validate:"min=1"`
		}

		val := validator.New(validator.WithRequiredStructEnabled())
		validationErrs := val.Struct(validationError{})

		w := httptest.NewRecorder()

		writeErrorResponse(w, validationErrs)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.JSONEq(
			t,
			`["Field validation for 'Name' failed on the 'required' tag",`+
				`"Field validation for 'Age' failed on the 'min' tag"]`,
			w.Body.String(),
		)
	})
}

func Test_writeResponse(t *testing.T) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		require.NoError(t, writeResponse(w, int(http.StatusOK), httpError{StatusCode: http.StatusOK, Message: "OK"}))
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.JSONEq(t, `{"StatusCode":200,"Message":"OK"}`, w.Body.String())
	})
	t.Run("default_status", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		require.NoError(t, writeResponse(w, 0, httpError{StatusCode: http.StatusOK, Message: "OK"}))
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
		assert.JSONEq(t, `{"StatusCode":200,"Message":"OK"}`, w.Body.String())
	})
	t.Run("marshal_error", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()

		type chanType struct {
			C chan int
		}

		err := writeResponse(w, int(http.StatusOK), chanType{C: make(chan int)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "json: unsupported type: chan int")
	})
}

func TestCreateHandler(t *testing.T) {
	t.Parallel()
	t.Run("request_type_as_struct_should_panics", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			_ = CreateHandler(
				func(r *http.Request) (sample.Request, error) {
					return sample.Request{}, nil
				},
				nil,
				func(ctx context.Context, req sample.Request) (sample.Response, int, error) {
					return sample.Response{}, http.StatusOK, nil
				})
		})
	})
	t.Run("parsing_error_should_return_bad_request", func(t *testing.T) {
		t.Parallel()

		handler := CreateHandler(
			func(r *http.Request) (*sample.Request, error) {
				return nil, errors.New("parsing error")
			},
			nil,
			func(ctx context.Context, req *sample.Request) (sample.Response, int, error) {
				return sample.Response{}, http.StatusOK, nil
			})
		w := httptest.NewRecorder()
		handler(w, httptest.NewRequest(http.MethodGet, "/", nil))
		assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
	})
}
