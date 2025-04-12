package server

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"time"
)

func getRemainingTime(server map[string]interface{}) time.Duration {
	creationDate, _, _ := unstructured.NestedString(server, "metadata", "creationTimestamp")
	createdTime, _ := time.Parse(time.RFC3339, creationDate)

	expireAfter, _, _ := unstructured.NestedInt64(server, "spec", "expireAfter")
	expireTime := createdTime.Add(time.Duration(expireAfter * int64(time.Millisecond)))

	return expireTime.Sub(time.Now())
}
