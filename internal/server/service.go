package server

import (
	"errors"
	"fmt"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
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

func (s *Service) getActiveServers() ([]server.Response, error) {
	activeServerResources, err := s.clusterService.GetActiveServers()
	if err != nil {
		return nil, err
	}

	activeServers := make([]server.Response, 0)
	for _, item := range activeServerResources.Items {
		tcpRoute, _, erro := unstructured.NestedString(item.Object, "spec", "routeName")
		if erro != nil {
			return nil, erro
		}
		objectStatus, _, _ := unstructured.NestedString(item.Object, "status", "status")

		var status string
		if objectStatus != "" {
			status = objectStatus
		} else {
			status = "not ready"
		}

		activeServers = append(activeServers, server.Response{
			Name:          item.GetName(),
			IP:            "minecraft.blamedevs.com:" + strconv.Itoa(s.getPortForTCPRoute(tcpRoute)),
			RemainingTime: getRemainingTime(item.Object).Round(time.Minute).String(),
			Status:        status,
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

func (s *Service) provisionServer(provisionRequest server.Request) (server.Response, error) {
	activeRoutes, err := s.clusterService.GetActiveRoutes()
	if err != nil {
		return server.Response{}, err
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
		return server.Response{}, errors.New("no TCP routes available")
	}

	activeServers, err := s.clusterService.GetActiveServers()
	if err != nil {
		return server.Response{}, err
	}
	for _, item := range activeServers.Items {
		if item.GetName() == provisionRequest.Name {
			return server.Response{}, errors.New("server with this name already exists")
		}
	}

	var route string
	for key := range availableRoutes {
		route = key
		break
	}

	serverRequest := cluster.ServerRequest{
		Name:      provisionRequest.Name,
		Image:     "ghcr.io/thijmengthn/papermc:latest",
		RouteName: route,
		Env: map[string]string{
			"EULA": "true",
		},
		ExpireAfter: provisionRequest.ExpireAfter,
	}
	err = s.clusterService.CreateServerInCluster(serverRequest)
	if err != nil {
		return server.Response{}, err
	}

	res := server.Response{
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
		remainingTime := getRemainingTime(server.Object)
		if remainingTime != remainingTime.Abs() {
			fmt.Println("Server " + server.GetName() + " expired initiating shutdown")
			s.clusterService.DeleteServer(server.GetName())
			serversExpired = append(serversExpired, server.GetName())
		} else {
			fmt.Println("Server " + server.GetName() + " remaining time " + remainingTime.String())
		}
	}

	return expiredServers{Names: serversExpired}, err
}
