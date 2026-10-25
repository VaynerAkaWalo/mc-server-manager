package definition

import "fmt"

type InvalidDefinitionError struct {
	Message string
}

func (e InvalidDefinitionError) Error() string {
	return e.Message
}

func invalidf(format string, args ...any) error {
	return InvalidDefinitionError{Message: fmt.Sprintf(format, args...)}
}
