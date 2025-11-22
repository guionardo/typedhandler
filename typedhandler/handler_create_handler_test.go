package typedhandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/guionardo/typedhandler/typedhandler"
	"github.com/stretchr/testify/assert"
)

func parseRequest(request *http.Request) (*typedhandler.Request, error) {
	var instance typedhandler.Request

	err := json.NewDecoder(request.Body).Decode(&instance)

	return &instance, err
}

func serviceRun(ctx context.Context, request *typedhandler.Request) (typedhandler.Response, int, error) {
	return typedhandler.Response{Message: request.Name}, http.StatusOK, nil
}
func serviceRunNormal(ctx context.Context, request *typedhandler.RequestNormal) (typedhandler.Response, int, error) {
	return typedhandler.Response{Message: request.Name}, http.StatusOK, nil
}

func TestCreateHandler(t *testing.T) {
	t.Parallel()
	t.Run("no_error", func(t *testing.T) {
		t.Parallel()

		typedhandler.ResetCreatedInstances()

		parser := typedhandler.CreateParser[*typedhandler.Request]()
		handler := typedhandler.CreateHandler(parser, serviceRun)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
		assert.Equal(t, http.StatusOK, response.Result().StatusCode)
		assert.JSONEq(t, "{\"Message\":\"John Doe\"}", response.Body.String())

		response = httptest.NewRecorder()
		request, _ = http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"Mary Doe"}`)))
		handler(response, request)
		assert.Equal(t, http.StatusOK, response.Result().StatusCode)
		assert.JSONEq(t, "{\"Message\":\"Mary Doe\"}", response.Body.String())
	})
	t.Run("user_parser_no_error", func(t *testing.T) {
		t.Parallel()

		typedhandler.ResetCreatedInstances()

		handler := typedhandler.CreateHandler(parseRequest, serviceRun)
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

	typedhandler.ResetCreatedInstances()

	parser := typedhandler.CreateParser[*typedhandler.Request]()
	for b.Loop() {
		handler := typedhandler.CreateHandler(parser, serviceRun)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
	}

	b.ReportMetric(float64(typedhandler.CreatedInstances()), "instances")
}

func easyJsonPoolDisabled(b *testing.B) {
	typedhandler.PoolEnabled = false

	typedhandler.ResetCreatedInstances()

	parser := typedhandler.CreateParser[*typedhandler.Request]()
	for b.Loop() {
		handler := typedhandler.CreateHandler(parser, serviceRun)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
	}

	b.ReportMetric(float64(typedhandler.CreatedInstances()), "instances")
}

func normalJsonPoolEnabled(b *testing.B) {
	typedhandler.PoolEnabled = true

	typedhandler.ResetCreatedInstances()

	parser := typedhandler.CreateParser[*typedhandler.RequestNormal]()
	for b.Loop() {
		handler := typedhandler.CreateHandler(parser, serviceRunNormal)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
	}

	b.ReportMetric(float64(typedhandler.CreatedInstances()), "instances")
}

func normalJsonPoolDisabled(b *testing.B) {
	typedhandler.PoolEnabled = false

	typedhandler.ResetCreatedInstances()

	parser := typedhandler.CreateParser[*typedhandler.RequestNormal]()
	for b.Loop() {
		handler := typedhandler.CreateHandler(parser, serviceRunNormal)
		request, _ := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(`{"name":"John Doe"}`)))
		response := httptest.NewRecorder()
		handler(response, request)
	}

	b.ReportMetric(float64(typedhandler.CreatedInstances()), "instances")
}
