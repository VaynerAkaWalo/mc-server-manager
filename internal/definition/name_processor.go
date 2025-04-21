package definition

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

func processName(ctx context.Context, name string) (string, error) {
	name = strings.ReplaceAll(name, " ", "-")

	pattern := regexp.MustCompile("[^a-zA-Z0-9-]+")

	sanitizedString := pattern.ReplaceAllString(name, "")

	if len(sanitizedString) < 5 || len(sanitizedString) > 25 {
		return "", fmt.Errorf("the server name must be between 5 and 25 characters long")
	}

	return sanitizedString, nil
}
