package definition

import (
	"context"
	"errors"
	"testing"
)

func TestValidateNotDowngrade(t *testing.T) {
	tests := []struct {
		name      string
		played    string
		requested string
		wantErr   bool
	}{
		{"same version", "1.21.3", "1.21.3", false},
		{"newer version", "1.21.3", "1.21.10", false},
		{"older version", "1.21.10", "1.21.3", true},
		{"never played", "", "1.21.3", false},
		{"played version no longer listed", "1.8", "1.21.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotDowngrade(context.Background(), testVersions, tt.played, tt.requested)

			var invalid InvalidDefinitionError
			if tt.wantErr != errors.As(err, &invalid) {
				t.Fatalf("got error %v", err)
			}
		})
	}
}
