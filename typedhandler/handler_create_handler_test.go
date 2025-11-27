package typedhandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/guionardo/typedhandler/examples/sample"
	"github.com/guionardo/typedhandler/typedhandler"
	"github.com/stretchr/testify/assert"
)

func parseRequest(request *http.Request) (*sample.Request, error) {
	var instance sample.Request

	err := json.NewDecoder(request.Body).Decode(&instance)

	return &instance, err
}

func serviceRun(ctx context.Context, request *sample.Request) (sample.Response, int, error) {
	return sample.Response{Message: request.Name}, http.StatusOK, nil
}
func serviceRunNormal(ctx context.Context, request *sample.RequestNormal) (sample.Response, int, error) {
	return sample.Response{Message: request.Name}, http.StatusOK, nil
}

func TestCreateHandler(t *testing.T) {
	t.Parallel()
	t.Run("no_error", func(t *testing.T) {
		t.Parallel()

		handler := typedhandler.CreateSimpleHandler(serviceRun)
		request, _ := http.NewRequest(
			"POST",
			"/",
			bytes.NewBuffer([]byte(`{"name":"John Doe","email":"john@doe.com"}`)),
		)
		response := httptest.NewRecorder()
		handler(response, request)
		assert.Equal(t, http.StatusOK, response.Result().StatusCode)
		assert.JSONEq(t, "{\"Message\":\"John Doe\"}", response.Body.String())

		response = httptest.NewRecorder()
		request, _ = http.NewRequest(
			"POST",
			"/",
			bytes.NewBuffer([]byte(`{"name":"Mary Doe","email":"user@host.net"}`)),
		)
		handler(response, request)
		assert.Equal(t, http.StatusOK, response.Result().StatusCode)
		assert.JSONEq(t, "{\"Message\":\"Mary Doe\"}", response.Body.String())
	})
	t.Run("user_parser_no_error", func(t *testing.T) {
		t.Parallel()

		handler := typedhandler.CreateHandler(parseRequest, nil, serviceRun)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
		assert.Equal(t, http.StatusOK, response.Result().StatusCode)
		assert.JSONEq(t, "{\"Message\":\"John Doe\"}", response.Body.String())
	})
}

func BenchmarkCreateHandler(b *testing.B) {
	b.ReportAllocs()
	b.Run("easyjson_pool_enabled", easyjsonPoolEnabled)
	b.Run("easyjson_pool_disabled", easyJsonPoolDisabled)
	b.Run("normal_json_pool_enabled", normalJsonPoolEnabled)
	b.Run("normal_json_pool_disabled", normalJsonPoolDisabled)
}

func easyjsonPoolEnabled(b *testing.B) {
	typedhandler.PoolEnabled = true

	parser, doneFunc := typedhandler.CreateParser[*sample.Request]()
	for b.Loop() {
		handler := typedhandler.CreateHandler(parser, doneFunc, serviceRun)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
	}

}

func easyJsonPoolDisabled(b *testing.B) {
	typedhandler.PoolEnabled = false

	parser, doneFunc := typedhandler.CreateParser[*sample.Request]()
	for b.Loop() {
		handler := typedhandler.CreateHandler(parser, doneFunc, serviceRun)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
	}

}

func normalJsonPoolEnabled(b *testing.B) {
	typedhandler.PoolEnabled = true

	parser, doneFunc := typedhandler.CreateParser[*sample.RequestNormal]()
	for b.Loop() {
		handler := typedhandler.CreateHandler(parser, doneFunc, serviceRunNormal)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
	}

}

func normalJsonPoolDisabled(b *testing.B) {
	typedhandler.PoolEnabled = false

	parser, doneFunc := typedhandler.CreateParser[*sample.RequestNormal]()
	for b.Loop() {
		handler := typedhandler.CreateHandler(parser, doneFunc, serviceRunNormal)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
	}

}
