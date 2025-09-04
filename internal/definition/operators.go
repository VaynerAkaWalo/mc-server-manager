package definition

import (
	"context"
	"strings"
)

const (
	AdminOnline  = "554156b6-2a93-3fe2-a63e-45f4fa95ec35"
	AdminOffline = "510bd8128b1b4a5cacd80143a76cab51"
)

func customizeOperators(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	operators := strings.Join([]string{AdminOnline, AdminOffline}, ",")
	if definition.Options[OPS] == "" {
		definition.Options[OPS] = operators
	} else {
		definition.Options[OPS] = operators + "," + definition.Options[OPS]
	}

	return definition, nil
}
