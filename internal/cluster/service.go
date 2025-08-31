package cluster

import (
	"context"
	serversv1alpha1 "github.com/VaynerAkaWalo/mc-server-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"log/slog"
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

func serverGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "servers.blamedevs.com",
		Version:  "v1alpha1",
		Resource: "mcservers",
	}
}

func (s *Service) GetActiveServers() (*unstructured.UnstructuredList, error) {
	return s.dynamicClient.Resource(serverGVR()).List(context.TODO(), metav1.ListOptions{})
}

func (s *Service) DeployServerSpec(serverRequest serversv1alpha1.McServerSpec) error {
	server := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "servers.blamedevs.com/v1alpha1",
			"kind":       "McServer",
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app.kubernetes.io/name":       serverRequest.Name,
					"app.kubernetes.io/managed-by": "kustomize",
				},
				"name":      serverRequest.Name,
				"namespace": "minecraft-server",
			},
			"spec": map[string]interface{}{
				"name":         serverRequest.Name,
				"image":        serverRequest.Image,
				"env":          serverRequest.Env,
				"expireAfter":  serverRequest.ExpireAfter,
				"memory":       serverRequest.Memory,
				"instanceType": "shared",
			},
			"status": map[string]interface{}{
				"status":      "new",
				"startedTime": "",
			},
		},
	}

	slog.Info("Creating new server %s, and image %s", serverRequest.Name, serverRequest.Image)
	_, err := s.dynamicClient.Resource(serverGVR()).Namespace("minecraft-server").Create(context.TODO(), server, metav1.CreateOptions{})
	if err != nil {
		slog.Info("Failed to create server " + err.Error())
		return err
	}

	return nil
}

func (s *Service) DeleteServer(serverName string) {
	err := s.dynamicClient.Resource(serverGVR()).Namespace("minecraft-server").Delete(context.TODO(), serverName, metav1.DeleteOptions{})
	if err != nil {
		slog.Error("Failed to delete server" + serverName)
	}
}
