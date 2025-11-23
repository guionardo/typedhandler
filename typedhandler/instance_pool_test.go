package typedhandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	noClearingRequest struct{}
	resettableRequest struct {
		Value string `query:"value"`
	}
	normalRequest struct {
		Value string `query:"value"`
	}
)

func Test_createResetFunc(t *testing.T) {
	t.Parallel()

	t.Run("pool_for_noclearing_request", testNoClearingRequest)
	t.Run("pool_for_resettable_request", testResettableRequest)
	t.Run("pool_for_normal_request", testNormalRequest)
}

func testNoClearingRequest(t *testing.T) {
	t.Parallel()

	infos := getParserInfos[noClearingRequest]()
	resetFunc := createResetFunc[noClearingRequest](infos)

	assert.NotPanics(t, func() {
		resetFunc(noClearingRequest{})
	})
}

func testResettableRequest(t *testing.T) {
	t.Parallel()

	infos := getParserInfos[*resettableRequest]()
	resetFunc := createResetFunc[*resettableRequest](infos)
	instance := &resettableRequest{Value: "test"}
	resetFunc(instance)
	assert.Empty(t, instance.Value)
}

func testNormalRequest(t *testing.T) {
	t.Parallel()

	infos := getParserInfos[*normalRequest]()
	resetFunc := createResetFunc[*normalRequest](infos)
	instance := &normalRequest{Value: "test"}
	resetFunc(instance)
	assert.Empty(t, instance.Value)

	instance = nil
	resetFunc(instance)
	assert.Nil(t, instance)
}

func (r *resettableRequest) Reset() {
	r.Value = ""
}

func TestMustBeAPointer(t *testing.T) {
	t.Parallel()
	t.Run("must_be_a_pointer_panics_for_non_pointer", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() {
			MustBeAPointer[noClearingRequest]()
		})
	})
	t.Run("must_be_a_pointer_does_not_panic_for_pointer", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			MustBeAPointer[*noClearingRequest]()
		})
	})
}

func TestNewInstancePool(t *testing.T) {
	t.Parallel()

	resetFunc := createResetFunc[*normalRequest](getParserInfos[*normalRequest]())
	pool := NewInstancePool(resetFunc)

	p1 := pool.Get()
	p1.Value = "test-1"

	p2 := pool.Get()
	p2.Value = "test-2"
	assert.NotSame(t, p1, p2, "expected to get different instances from the pool")

	pool.Put(p1)
	p3 := pool.Get()
	assert.Same(t, p1, p3, "expected to get the same instance from the pool")
	assert.Empty(t, p3.Value, "expected reset function to clear the Value field")
}
