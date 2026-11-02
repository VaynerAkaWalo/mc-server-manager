package server

import (
	"errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"net"
)

func serverAddress(serverName string) string {
	return serverName + ".mc-server.blamedevs.com"
}

func clusterAddress(serverName string) string {
	return net.JoinHostPort(serverName+".minecraft-server.svc", "25565")
}

func rejectedByCluster(err error) bool {
	var status apierrors.APIStatus
	if !errors.As(err, &status) {
		return false
	}

	code := status.Status().Code
	return code >= 400 && code < 500 && code != 408 && code != 429
}
