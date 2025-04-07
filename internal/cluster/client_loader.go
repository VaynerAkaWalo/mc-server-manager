package cluster

import (
	"flag"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

type ClientLoader struct {
	config rest.Config
}

func CreateClientLoader() ClientLoader {
	config, err := localConfig()
	if err != nil {
		log.Fatal("Failed to load config")
	}

	return ClientLoader{
		config: *config,
	}
}

func (l *ClientLoader) Client() (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(&l.config)
}

func (l *ClientLoader) DynamicClient() (*dynamic.DynamicClient, error) {
	return dynamic.NewForConfig(&l.config)
}

func localConfig() (*rest.Config, error) {
	kubeconfig := flag.String("kubeconfig", "./kubeconfig", "")
	flag.Parse()

	return clientcmd.BuildConfigFromFlags("", *kubeconfig)
}

func inClusterConfig() (*rest.Config, error) {
	return rest.InClusterConfig()
}
