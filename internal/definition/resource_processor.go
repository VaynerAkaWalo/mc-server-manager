package definition

import "context"

func processQuota(ctx context.Context, quota *ResourceQuota) (*ResourceQuota, error) {
	if quota.CpuRequest == 0 {
		quota.CpuRequest = DefaultQuota.CpuRequest
	}

	if quota.CpuLimit == 0 {
		quota.CpuLimit = DefaultQuota.CpuLimit
	}

	if quota.CpuRequest > quota.CpuLimit {
		quota.CpuRequest = DefaultQuota.CpuRequest
	}

	if quota.MemoryInMb == 0 {
		quota.MemoryInMb = 7800
	}

	return quota, nil
}
