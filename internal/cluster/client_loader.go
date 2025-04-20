package cluster

import (
	"flag"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type ClientLoader struct {
	config rest.Config
}

func CreateClientLoader() ClientLoader {
	var config *rest.Config
	if os.Getenv("env") != "prod" {
		config, _ = localConfig()
	} else {
		config, _ = rest.InClusterConfig()
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
