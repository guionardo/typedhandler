package typedhandler

type (
	// Resettable represents a struct that can reset its fields to default values
	Resettable interface {
		Reset()
	}

	Validatable interface {
		Validate() error
	}
	RequestSchema  any
	ResponseSchema any
)
