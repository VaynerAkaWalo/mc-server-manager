package definition

import (
	"context"
	"fmt"
	"regexp"
	"time"
)

const (
	MinDurationForSave = 2 * time.Hour
	AllowedCharsRegex  = "[^a-zA-Z0-9 ]+"
)

func saveCustomizer(ctx context.Context, definition ServerDefinition) (ServerDefinition, error) {
	if definition.SaveId == "" && definition.Duration >= MinDurationForSave {
		saveId := fmt.Sprintf("%s-%v", definition.Name, time.Now().Format(time.RFC3339))
		definition.SaveId = getSanitizedName(saveId)
	}

	return definition, nil
}

func getSanitizedName(name string) string {
	regex, _ := regexp.Compile(AllowedCharsRegex)
	return regex.ReplaceAllString(name, "-")
}
