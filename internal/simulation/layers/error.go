package layers

const (
	CodeGenerationError     = "code_generation_error"
	CodeClassificationError = "code_classification_error"
)

type PipelineError struct {
	Code    string
	Message string
	Cause   error
}

func (e *PipelineError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *PipelineError) Unwrap() error {
	return e.Cause
}
