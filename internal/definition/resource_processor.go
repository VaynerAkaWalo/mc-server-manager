package definition

import "context"

func processQuota(ctx context.Context, quota *ResourceQuota) (*ResourceQuota, error) {
	if quota.MemoryInMb == 0 {
		quota.MemoryInMb = 7800
	}

	return quota, nil
}
