package typedhandler

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/guionardo/typedhandler/examples/sample"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	RequestInvalidHeader struct {
		AuthToken int `header:"Authorization"`
	}
	StructWithUnsettableField struct {
		UnsettableField *struct{ name string } `path:"unsettable_field"`
	}
	noClearingRequest struct{}
	resettableRequest struct {
		Value string `query:"value"`
	}
	normalRequest struct {
		Value string `query:"value"`
	}
	requestWithBody struct {
		Body body `body:""`
	}
	body struct {
		BodyField string `json:"body_field"`
	}
)

func (r *resettableRequest) Reset() {
	r.Value = ""
}

func (r *requestWithBody) GetBodyField() any {
	return &r.Body
}

func TestSchemaHelper(t *testing.T) {
	t.Parallel()
	t.Run("GetSchemaHelper", func(t *testing.T) {
		t.Parallel()

		sh := GetSchemaHelper[*sample.Request]()
		require.NotNil(t, sh)

		sh2 := GetSchemaHelper[*sample.Request]()
		require.Same(t, sh, sh2)
	})
	t.Run("Schema with invalid type for header", func(t *testing.T) {
		t.Parallel()
		require.Panics(t, func() {
			_ = GetSchemaHelper[*RequestInvalidHeader]()
		})
	})
	t.Run("Schema with unsettable field", func(t *testing.T) {
		t.Parallel()
		t.Skip("Can't find a unsettable field type for now")
		require.Panics(t, func() {
			hlp := GetSchemaHelper[*StructWithUnsettableField]()
			_ = hlp
		})
	})
}

func TestSchemaHelper_InstancePool(t *testing.T) {
	t.Parallel()

	sh := GetSchemaHelper[*sample.Request]()
	req1 := sh.GetInstance()
	req1.Name = "Test1"

	req2 := sh.GetInstance()
	require.NotSame(t, req1, req2)

	sh.PutInstance(req1)

	req1Again := sh.GetInstance()
	require.Same(t, req1, req1Again)
	require.Empty(t, req1Again.Name)
}

func TestSchemaHelper_CreateResetFunc(t *testing.T) {
	t.Parallel()
	t.Run("noclearing_request", func(t *testing.T) {
		t.Parallel()

		helper := GetSchemaHelper[*noClearingRequest]()
		instance := &noClearingRequest{}

		assert.NotPanics(t, func() {
			helper.ResetFunc(instance)
		})
	})
	t.Run("resettable_request", func(t *testing.T) {
		t.Parallel()

		helper := GetSchemaHelper[*resettableRequest]()
		instance := &resettableRequest{Value: "test"}
		helper.ResetFunc(instance)
		assert.Empty(t, instance.Value)
	})
	t.Run("normal_request", func(t *testing.T) {
		t.Parallel()

		helper := GetSchemaHelper[*normalRequest]()
		instance := &normalRequest{Value: "test"}
		helper.ResetFunc(instance)
		assert.Empty(t, instance.Value)

		instance = nil
		helper.ResetFunc(instance)
		assert.Nil(t, instance)
	})
}

func Test_parseBodyInstance(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"tester"}`))
	require.NoError(t, err)

	type parserType struct {
		Name string `json:"name"`
	}

	var responseBody parserType

	err = parseBodyInstance(req, &responseBody)
	require.NoError(t, err)
	assert.Equal(t, "tester", responseBody.Name)
}

func Test_parseBodyField(t *testing.T) {
	t.Parallel()
	t.Run("no_get_body_fielder", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"tester"}`))
		require.NoError(t, err)

		type parserType struct {
			Name string `json:"name"`
		}

		var responseBody parserType

		err = parseBodyField(req, &responseBody)
		require.Error(t, err)
	})
	t.Run("with_get_body_fielder", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"body_field":"tester"}`))
		require.NoError(t, err)

		rwf := requestWithBody{}
		err = parseBodyField(req, &rwf)
		require.NoError(t, err)
		assert.Equal(t, "tester", rwf.Body.BodyField)
	})
}
