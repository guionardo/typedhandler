package typedhandler

import "net/http"

type (
	// Resettable represents a struct that can reset its fields to default values
	Resettable interface {
		Reset()
	}

	Validatable interface {
		Validate() error
	}

	PreParseable interface {
		PreParse(r *http.Request) error
	}

	HttpError interface {
		error
		Status() int
	}

	HttpJsonError interface {
		HttpError
		Json() []byte
	}

	RequestSchema  any
	ResponseSchema any
)
