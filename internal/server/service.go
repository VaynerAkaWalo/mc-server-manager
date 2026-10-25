package server

import (
	"context"
	"errors"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	"github.com/google/uuid"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log/slog"
	"time"
)

var ErrNameInUse = errors.New("a server with this name already exists")

type serverStore interface {
	CreateServer(ctx context.Context, server store.NewServer) (store.ServerRecord, error)
	DeleteServer(ctx context.Context, id uuid.UUID) error
	CountServers(ctx context.Context) (store.ServerCounts, error)
}

type Service struct {
	clusterService cluster.Service
	store          serverStore
}

func CreateServerService(clusterService cluster.Service, store serverStore) Service {
	return Service{
		clusterService: clusterService,
		store:          store,
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
			IP:            serverAddress(item.GetName()),
			RemainingTime: getRemainingTime(item.Object).Round(time.Minute).String(),
			Status:        status,
		})
	}

	return activeServers, nil
}

func (s *Service) getStats(ctx context.Context) (stats, error) {
	counts, err := s.store.CountServers(ctx)
	if err != nil {
		return stats{}, err
	}

	return stats{Online: counts.Online, Created: counts.Created}, nil
}

func (s *Service) provisionServer(provisionRequest server.Request) (server.Response, error) {
	activeServers, err := s.clusterService.GetActiveServers()
	if err != nil {
		return server.Response{}, err
	}
	for _, item := range activeServers.Items {
		if item.GetName() == provisionRequest.Name {
			return server.Response{}, ErrNameInUse
		}
	}

	serverDefinition := definition.ServerDefinition{
		Name:     provisionRequest.Name,
		Options:  provisionRequest.Opts,
		Tier:     definition.Tier(provisionRequest.Tier),
		Preset:   provisionRequest.Preset,
		Duration: time.Duration(provisionRequest.Duration) * time.Millisecond,
	}

	serverSpec, err := definition.TranslateDefinition(context.TODO(), serverDefinition)
	if err != nil {
		return server.Response{}, err
	}

	record, err := s.store.CreateServer(context.TODO(), store.NewServer{
		Name:       serverSpec.Name,
		Hostname:   serverAddress(serverSpec.Name),
		Definition: serverDefinition,
		ExpiresAt:  time.Now().Add(serverDefinition.Duration),
	})
	if err != nil {
		return server.Response{}, err
	}

	err = s.clusterService.DeployServerSpec(*serverSpec)
	if err != nil {
		alreadyExists := apierrors.IsAlreadyExists(err)
		invalid := apierrors.IsInvalid(err)
		if rejectedByCluster(err) {
			if deleteErr := s.store.DeleteServer(context.TODO(), record.ID); deleteErr != nil {
				slog.Error("Failed to remove record of server that was not deployed " + record.ID.String() + ": " + deleteErr.Error())
			}
		}
		if alreadyExists {
			return server.Response{}, ErrNameInUse
		}
		if invalid {
			return server.Response{}, definition.InvalidDefinitionError{Message: "the cluster rejected the server definition, check the server name"}
		}
		return server.Response{}, err
	}

	res := server.Response{
		Name: provisionRequest.Name,
		IP:   serverAddress(provisionRequest.Name),
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
			slog.Info("Server " + server.GetName() + " expired initiating shutdown")
			s.clusterService.DeleteServer(server.GetName())
			serversExpired = append(serversExpired, server.GetName())
		} else {
			slog.Info("Server " + server.GetName() + " remaining time " + remainingTime.String())
		}
	}

	return expiredServers{Names: serversExpired}, err
}
