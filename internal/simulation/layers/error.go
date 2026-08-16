package layers

type GeneratorError struct {
	Message string
	Cause   error
}

func (e *GeneratorError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *GeneratorError) Unwrap() error {
	return e.Cause
}
