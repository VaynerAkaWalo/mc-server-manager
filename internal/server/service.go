package server

import (
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strconv"
)

type Service struct {
	clusterService cluster.Service
}

func CreateServerService(clusterService cluster.Service) Service {
	return Service{
		clusterService: clusterService,
	}
}

func (s *Service) getActiveServers() ([]response, error) {
	activeServerResources, err := s.clusterService.ListActiveServers()
	if err != nil {
		return nil, err
	}

	var activeServers []response
	for _, item := range activeServerResources.Items {
		tcpRoute, _, erro := unstructured.NestedString(item.Object, "spec", "routeName")
		if erro != nil {
			return nil, erro
		}

		activeServers = append(activeServers, response{
			Name: item.GetName(),
			IP:   "blamedevs.com:" + strconv.Itoa(s.getPortForTCPRoute(tcpRoute)),
		})
	}

	return activeServers, nil
}

func (s *Service) getPortForTCPRoute(tcpRoute string) int {
	portMap := map[string]int{
		"tcp-1": 25570,
		"tcp-2": 25571,
		"tcp-3": 25572,
	}

	return portMap[tcpRoute]
}
