package go_config_reader

import (
	"github.com/qri-io/jsonschema"
)

//------------------------------------------------------------------------------

type ValidationError struct {
	Failures []ValidationErrorFailure
}

type ValidationErrorFailure struct {
	Location string
	Message  string
}

//------------------------------------------------------------------------------

func NewValidationError(errors []jsonschema.KeyError) *ValidationError {
	err := ValidationError{
		Failures: make([]ValidationErrorFailure, len(errors)),
	}

	for idx, e := range errors {
		err.Failures[idx].Location = e.PropertyPath
		err.Failures[idx].Message = e.Message
	}

	return &err
}

func (e *ValidationError) Error() string {
	desc := ""
	if len(e.Failures) > 0 {
		desc = " / " + e.Failures[0].Message + " @ " + e.Failures[0].Location
	}
	return "unable to load configuration [validation failed" + desc + "]"
}

func (_ *ValidationError) Unwrap() error {
	return nil
}
