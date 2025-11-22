package typedhandler

import (
	"reflect"
	"sync"
	"sync/atomic"
)

type (
	// InstancePool is a pool of instances of type RIn to reuse memory and reduce allocations
	InstancePool[RIn RequestSchema] struct {
		pool      sync.Pool
		get       func() RIn
		put       func(RIn)
		resetFunc func(RIn)
	}
	// Resettable is an interface that can be used to reset the instance to its initial state
	Resettable interface {
		Reset()
	}
)

var (
	pools            sync.Map
	createdInstances atomic.Int64
	PoolEnabled      = true
)

func NewInstancePool[RIn RequestSchema](resetFunc func(RIn)) *InstancePool[RIn] {
	typeName := typeName[RIn]()
	if pool, ok := pools.Load(typeName); ok {
		return pool.(*InstancePool[RIn])
	}

	var pool = &InstancePool[RIn]{
		resetFunc: resetFunc,
	}

	if PoolEnabled {
		pool.pool = sync.Pool{
			New: func() any {
				return NewInstance[RIn](reflect.TypeFor[RIn]())
			},
		}
		pool.get = func() RIn {
			instance := pool.pool.Get().(RIn)
			pool.resetFunc(instance)

			return instance
		}
		pool.put = func(instance RIn) {
			pool.pool.Put(instance)
		}
	} else {
		pool.get = func() RIn {
			return NewInstance[RIn](reflect.TypeFor[RIn]())
		}
		pool.put = func(instance RIn) {}
	}

	pools.Store(typeName, pool)

	return pool
}

func (p *InstancePool[RIn]) Get() RIn {
	return p.pool.Get().(RIn)
}

func (p *InstancePool[RIn]) Put(instance RIn) {
	p.pool.Put(instance)
}

func NewInstance[RIn RequestSchema](dataType reflect.Type) (instance RIn) {
	createdInstances.Add(1)
	// Get the underlying struct type (strip pointer if dataType is a pointer)
	structType := dataType
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	// Check if RIn is a pointer type
	var zero RIn

	rInType := reflect.TypeOf(zero)

	if rInType.Kind() == reflect.Pointer {
		// RIn is a pointer type, create a pointer to the struct
		// reflect.New(structType) returns a pointer to structType
		ptr := reflect.New(structType)
		return ptr.Interface().(RIn)
	}
	// RIn is a value type, create a value instance
	structValue := reflect.New(structType).Elem()

	return structValue.Interface().(RIn)
}

func CreatedInstances() int64 {
	return createdInstances.Load()
}

func ResetCreatedInstances() {
	createdInstances.Store(0)
}

// createResetFunc creates a function to reset instance fields to zero values
func createResetFunc[RIn RequestSchema](infos *ParserInfos) func(RIn) {
	if (len(infos.headerFields) + len(infos.queryFields) + len(infos.pathFields)) == 0 {
		// Only create reset function if we have non-body fields that need clearing
		return func(RIn) {
			// NOOP
		}
	}

	structType := infos.dataType

	var zero RIn

	// Check if type implements Resettable interface (compile-time check)
	if _, ok := any(zero).(Resettable); ok {
		// Use the Resettable interface method - no reflection needed
		return func(instance RIn) {
			any(instance).(Resettable).Reset()
		}
	}

	rInType := reflect.TypeOf(zero)

	if rInType.Kind() == reflect.Pointer {
		// For pointer types, reset all fields
		return func(instance RIn) {
			ptrValue := reflect.ValueOf(instance)
			if ptrValue.IsNil() {
				return
			}

			structValue := ptrValue.Elem()
			zeroValue := reflect.Zero(structType)
			structValue.Set(zeroValue)
		}
	}
	// For value types, reset all fields
	return func(instance RIn) {
		instanceValue := reflect.ValueOf(&instance).Elem()
		zeroValue := reflect.Zero(structType)
		instanceValue.Set(zeroValue)
	}
}
