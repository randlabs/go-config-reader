package go_config_reader

//------------------------------------------------------------------------------

type ValidationError struct {
	Failures []ValidationErrorFailure
}

type ValidationErrorFailure struct {
	Location string
	Message  string
}

//------------------------------------------------------------------------------

func (v *ValidationError) Error() string {
	errDesc := ""
	if len(v.Failures) > 0 {
		errDesc = "  / " + v.Failures[0].Message + " @ " + v.Failures[0].Location
	}
	return "unable to load configuration [validation failed" + errDesc + "]"
}
