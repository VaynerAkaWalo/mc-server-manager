package definition

import "fmt"

// InvalidDefinitionError means the requested server is not allowed, as opposed to a failure of the system building it.
type InvalidDefinitionError struct {
	Message string
}

func (e InvalidDefinitionError) Error() string {
	return e.Message
}

func invalidf(format string, args ...any) error {
	return InvalidDefinitionError{Message: fmt.Sprintf(format, args...)}
}
