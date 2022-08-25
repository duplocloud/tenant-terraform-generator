package tfgenerator

type TFGeneratorError struct {
	ErrorMessage string
}

func (e *TFGeneratorError) Error() string {
	return e.ErrorMessage
}

func ThrowError(error string) error {
	return &TFGeneratorError{ErrorMessage: error}
}
