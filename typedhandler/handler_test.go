package typedhandler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	httpError struct {
		status  int
		message string
	}
	jsonError struct {
		httpError
	}
)

func (e httpError) Status() int {
	return e.status
}

func (e httpError) Error() string {
	return e.message
}

func (e jsonError) Json() []byte {
	return fmt.Appendf(nil, "{\"message\":\"%s\"}", e.message)
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
		writeErrorResponse(w, httpError{status: http.StatusBadRequest, message: "Bad Request"})
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.Equal(t, "Bad Request", w.Body.String())
	})
	t.Run("http_json_error", func(t *testing.T) {
		t.Parallel()

		w := httptest.NewRecorder()
		writeErrorResponse(w, jsonError{httpError: httpError{status: http.StatusBadRequest, message: "Bad Request"}})
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
		assert.JSONEq(t, "{\"message\":\"Bad Request\"}", w.Body.String())
	})
}
