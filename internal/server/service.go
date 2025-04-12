package server

import (
	"errors"
	"fmt"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strconv"
	"time"
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
	activeServerResources, err := s.clusterService.GetActiveServers()
	if err != nil {
		return nil, err
	}

	activeServers := make([]response, 0)
	for _, item := range activeServerResources.Items {
		tcpRoute, _, erro := unstructured.NestedString(item.Object, "spec", "routeName")
		if erro != nil {
			return nil, erro
		}

		activeServers = append(activeServers, response{
			Name: item.GetName(),
			IP:   "minecraft.blamedevs.com:" + strconv.Itoa(s.getPortForTCPRoute(tcpRoute)),
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

func (s *Service) provisionServer(provisionRequest request) (response, error) {
	activeRoutes, err := s.clusterService.GetActiveRoutes()
	if err != nil {
		return response{}, err
	}

	availableRoutes := map[string]bool{
		"tcp-1": true,
		"tcp-2": true,
		"tcp-3": true,
	}
	for _, item := range activeRoutes.Items {
		delete(availableRoutes, item.GetName())
	}

	if len(availableRoutes) < 1 {
		return response{}, errors.New("no TCP routes available")
	}

	activeServers, err := s.clusterService.GetActiveServers()
	if err != nil {
		return response{}, err
	}
	for _, item := range activeServers.Items {
		if item.GetName() == provisionRequest.Name {
			return response{}, errors.New("server with this name already exists")
		}
	}

	var route string
	for key := range availableRoutes {
		route = key
		break
	}

	server := cluster.ServerRequest{
		Name:      provisionRequest.Name,
		Image:     provisionRequest.Image,
		RouteName: route,
		Env: map[string]string{
			"EULA": "true",
		},
		ExpireAfter: provisionRequest.ExpireAfter,
	}
	err = s.clusterService.CreateServerInCluster(server)
	if err != nil {
		return response{}, err
	}

	res := response{
		Name: provisionRequest.Name,
		IP:   "minecraft.blamedevs.com:" + strconv.Itoa(s.getPortForTCPRoute(route)),
	}
	return res, nil
}

func (s *Service) shutdownExpiredServers() (expiredServers, error) {
	activeServerResources, err := s.clusterService.GetActiveServers()
	if err != nil {
		return expiredServers{}, err
	}

	serversExpired := make([]string, 0)
	for _, server := range activeServerResources.Items {
		creationDate, _, _ := unstructured.NestedString(server.Object, "metadata", "creationTimestamp")
		createdTime, _ := time.Parse(time.RFC3339, creationDate)

		expireAfter, _, _ := unstructured.NestedInt64(server.Object, "spec", "expireAfter")
		expireTime := createdTime.Add(time.Duration(expireAfter * int64(time.Millisecond)))

		expired := time.Now().After(expireTime)
		if expired == true {
			fmt.Println("Deleting server " + server.GetName())
			s.clusterService.DeleteServer(server.GetName())
			serversExpired = append(serversExpired, server.GetName())
		} else {
			fmt.Println("Server " + server.GetName() + " remaining time " + expireTime.Sub(time.Now()).String())
		}
	}

	return expiredServers{Names: serversExpired}, err
}
