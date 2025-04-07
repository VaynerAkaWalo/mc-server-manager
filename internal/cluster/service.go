package cluster

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Service struct {
	client        kubernetes.Clientset
	dynamicClient dynamic.DynamicClient
}

func CreateClusterService(client kubernetes.Clientset, dynamicClient dynamic.DynamicClient) Service {
	return Service{
		client:        client,
		dynamicClient: dynamicClient,
	}
}

func (s *Service) ListActiveServers() (*unstructured.UnstructuredList, error) {
	gvr := schema.GroupVersionResource{
		Group:    "servers.blamedevs.com",
		Version:  "v1alpha1",
		Resource: "mcservers",
	}

	return s.dynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
}
