package definition

import (
	"context"
	"fmt"
	"github.com/VaynerAkaWalo/mc-server-operator/api/v1alpha1"
	"strconv"
)

func TranslateDefinition(ctx context.Context, definition ServerDefinition) (*v1alpha1.McServerSpec, error) {
	sanitizedName, err := processName(ctx, definition.Name)
	if err != nil {
		return nil, err
	}

	environment, err := processEnvironment(ctx, definition.Options)
	if err != nil {
		return nil, err
	}

	quota, err := processQuota(ctx, &definition.Quota)
	if err != nil {
		return nil, err
	}

	expireAfter, err := processExpirationTime(ctx, definition.ExpireAfter)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.McServerSpec{
		Name:        sanitizedName,
		Image:       defaultImage,
		Env:         stringifyEnvironment(environment),
		CpuRequest:  strconv.Itoa(quota.CpuRequest),
		CpuLimit:    strconv.Itoa(quota.CpuLimit),
		Memory:      fmt.Sprintf("%dMi", quota.MemoryInMb),
		ExpireAfter: expireAfter,
	}, nil
}
