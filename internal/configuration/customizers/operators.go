package customizers

import (
	"context"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/configuration"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"strings"
)

const (
	AdminOnline  = "554156b6-2a93-3fe2-a63e-45f4fa95ec35"
	AdminOffline = "510bd8128b1b4a5cacd80143a76cab51"
)

type operatorsCustomizer struct {
	admins []string
}

func NewOperatorsCustomizer() configuration.Customizer {
	return &operatorsCustomizer{
		admins: []string{AdminOnline, AdminOffline},
	}
}

func (oc *operatorsCustomizer) Customize(ctx context.Context, env map[definition.Option]string) (map[definition.Option]string, error) {
	if len(oc.admins) == 0 {
		return env, nil
	}

	operators := strings.Join(oc.admins, ",")
	if env[definition.OPS] == "" {
		env[definition.OPS] = operators
	} else {
		env[definition.OPS] = operators + "," + env[definition.OPS]
	}

	return env, nil
}
