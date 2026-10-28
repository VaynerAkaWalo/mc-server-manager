package definition

import (
	"context"
	"testing"
)

func TestTranslateTier(t *testing.T) {
	tests := []struct {
		tier       Tier
		wantMemory string
		wantErr    bool
	}{
		{Guest, MemorySmall, false},
		{"iron", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			instance, err := translateTier(context.Background(), tt.tier)

			if (err != nil) != tt.wantErr || instance.memory != tt.wantMemory {
				t.Fatalf("got %+v and error %v", instance, err)
			}
		})
	}
}
