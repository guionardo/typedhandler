package typedhandler_test

import (
	"reflect"
	"testing"

	"github.com/guionardo/typedhandler/typedhandler"
	"github.com/stretchr/testify/assert"
)

func TestNewInstance(t *testing.T) {
	t.Parallel()
	t.Run("get_instance_struct", func(t *testing.T) {
		t.Parallel()

		req := typedhandler.NewInstance[typedhandler.Request](reflect.TypeFor[typedhandler.Request]())
		assert.Empty(t, req.Name)
	})
	t.Run("get_instance_pointer", func(t *testing.T) {
		t.Parallel()

		req := typedhandler.NewInstance[*typedhandler.Request](reflect.TypeFor[*typedhandler.Request]())
		assert.NotNil(t, req)
	})
}
