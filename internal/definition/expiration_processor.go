package definition

import (
	"context"
	"fmt"
	"time"
)

const SIX_HOURS = time.Duration(6 * time.Hour)

func processExpirationTime(ctx context.Context, expireAfter int64) (int64, error) {
	if expireAfter <= 0 {
		return 0, fmt.Errorf("expire after cannot be less than or equal to zero")
	}

	if SIX_HOURS.Milliseconds() < expireAfter {
		return 0, fmt.Errorf("expiration time bigger than 6 hours are not supported")
	}

	return expireAfter, nil
}
