package definition

import (
	"context"
	"testing"
)

func TestCustomizeNameRejectsReservedNames(t *testing.T) {
	for _, name := range []string{"mc-proxy", "MC-Server-Manager", "mc server operator"} {
		if _, err := customizeName(context.Background(), ServerDefinition{Name: name}); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}

	if _, err := customizeName(context.Background(), ServerDefinition{Name: "my-server"}); err != nil {
		t.Errorf("my-server: %v", err)
	}
}
