package definition

import (
	"context"
	"fmt"
	"time"
)

const MinDurationForSave = 2 * time.Hour

func saveCustomizer(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.SaveId == "" && definition.Duration >= MinDurationForSave {
		saveId := fmt.Sprintf("%s-%v", definition.Name, time.Now().Format(time.RFC3339))
		definition.SaveId = saveId
	}

	return definition, nil
}
