package server

import (
	"errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log/slog"
	"time"
)

func getRemainingTime(server map[string]interface{}) time.Duration {
	resourceCreationTime, _, _ := unstructured.NestedString(server, "metadata", "creationTimestamp")
	createdTime, _ := time.Parse(time.RFC3339, resourceCreationTime)

	expireAfter, _, _ := unstructured.NestedInt64(server, "spec", "expireAfter")
	expireIn := time.Duration(expireAfter * int64(time.Millisecond))

	expireTimeByCreation := createdTime.Add(expireIn)

	serverStartedTime, _, _ := unstructured.NestedString(server, "status", "startedTime")
	if serverStartedTime == "" && createdTime.Add(10*time.Minute).Before(time.Now()) {
		slog.Error("Started time is null and server was created over 10 minutes ago")
		return expireTimeByCreation.Sub(time.Now())
	} else if serverStartedTime == "" {
		slog.Warn("Server is still not ready")
		return expireIn
	}
	startedTime, _ := time.Parse(time.RFC3339, serverStartedTime)
	expireTimeByStartedTime := startedTime.Add(expireIn)

	return expireTimeByStartedTime.Sub(time.Now())
}

func serverAddress(serverName string) string {
	return serverName + ".mc-server.blamedevs.com"
}

// Timeouts and connection failures may still have created the resource, so only an explicit client error proves it was not.
func rejectedByCluster(err error) bool {
	var status apierrors.APIStatus
	if !errors.As(err, &status) {
		return false
	}

	code := status.Status().Code
	return code >= 400 && code < 500 && code != 408 && code != 429
}
