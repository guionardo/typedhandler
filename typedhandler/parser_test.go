package typedhandler_test

import (
	"reflect"
	"testing"

	"github.com/guionardo/typedhandler/examples/sample"
	"github.com/guionardo/typedhandler/typedhandler"
	"github.com/stretchr/testify/assert"
)

func TestNewInstance(t *testing.T) {
	t.Parallel()
	t.Run("get_instance_struct", func(t *testing.T) {
		t.Parallel()

		req := typedhandler.NewInstance[sample.Request](reflect.TypeFor[sample.Request]())
		assert.Empty(t, req.Name)
	})
	t.Run("get_instance_pointer", func(t *testing.T) {
		t.Parallel()

		req := typedhandler.NewInstance[*sample.Request](reflect.TypeFor[*sample.Request]())
		assert.NotNil(t, req)
	})
}
