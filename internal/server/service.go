package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	definition "github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
		objectStatus, _, _ := unstructured.NestedString(item.Object, "status", "status")

		var status string
		if objectStatus != "" {
			status = objectStatus
		} else {
			status = "not ready"
		}

		activeServers = append(activeServers, server.Response{
			Name:          item.GetName(),
			IP:            item.GetName() + ".blamedevs.com",
			RemainingTime: getRemainingTime(item.Object).Round(time.Minute).String(),
			Status:        status,
		})
	}

	return activeServers, nil
}

func (s *Service) provisionServer(provisionRequest server.Request) (server.Response, error) {
	activeServers, err := s.clusterService.GetActiveServers()
	if err != nil {
		return server.Response{}, err
	}
	for _, item := range activeServers.Items {
		if item.GetName() == provisionRequest.Name {
			return server.Response{}, errors.New("server with this name already exists")
		}
	}

	serverDefinition := definition.ServerDefinition{
		Name:        provisionRequest.Name,
		Options:     provisionRequest.OPTS,
		Quota:       definition.DefaultQuota,
		ExpireAfter: provisionRequest.ExpireAfter,
	}

	serverSpec, err := definition.TranslateDefinition(context.TODO(), serverDefinition)
	if err != nil {
		return server.Response{}, err
	}

	err = s.clusterService.DeployServerSpec(*serverSpec)
	if err != nil {
		return server.Response{}, err
	}

	res := server.Response{
		Name: provisionRequest.Name,
		IP:   provisionRequest.Name + ".blamedevs.com",
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
