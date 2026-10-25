package server

import (
	"context"
	"errors"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"github.com/VaynerAkaWalo/mc-server-manager/pkg/server"
	serversv1alpha1 "github.com/VaynerAkaWalo/mc-server-operator/api/v1alpha1"
	"github.com/google/uuid"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"log/slog"
	"maps"
	"time"
)

var ErrNameInUse = errors.New("a server with this name already exists")

type serverStore interface {
	CreateServer(ctx context.Context, server store.NewServer) (store.ServerRecord, error)
	DeleteServer(ctx context.Context, id uuid.UUID) error
	CountServers(ctx context.Context) (store.ServerCounts, error)
	ListServers(ctx context.Context, clusterNames []string) ([]store.ServerRecord, error)
}

type clusterService interface {
	GetActiveServers() (*unstructured.UnstructuredList, error)
	DeployServerSpec(serverRequest serversv1alpha1.McServerSpec) error
	DeleteServer(serverName string, uid types.UID) error
}

type Service struct {
	clusterService clusterService
	store          serverStore
}

func CreateServerService(clusterService clusterService, store serverStore) Service {
	return Service{
		clusterService: clusterService,
		store:          store,
	}
}

func (s *Service) listServers(ctx context.Context) ([]server.Response, error) {
	resources, err := s.clusterService.GetActiveServers()
	if err != nil {
		return nil, err
	}

	resourcesByName := make(map[string]*unstructured.Unstructured, len(resources.Items))
	names := make([]string, 0, len(resources.Items))
	for i := range resources.Items {
		resourcesByName[resources.Items[i].GetName()] = &resources.Items[i]
		names = append(names, resources.Items[i].GetName())
	}

	records, err := s.store.ListServers(ctx, names)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	servers := make([]server.Response, 0, len(records))
	for _, record := range records {
		// Records are newest first, so a resource belongs to the first record with its name.
		resource := resourcesByName[record.Name]
		delete(resourcesByName, record.Name)

		response, err := observedResponse(record, resource, now)
		if err != nil {
			return nil, err
		}
		servers = append(servers, response)
	}

	return servers, nil
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
		Options:  toOptions(provisionRequest.Options),
		Tier:     definition.Tier(provisionRequest.Tier),
		Preset:   provisionRequest.Preset,
		Duration: time.Duration(provisionRequest.DurationMs) * time.Millisecond,
	}

	// Translation mutates options, so preserve the requested definition.
	translated := serverDefinition
	translated.Options = maps.Clone(serverDefinition.Options)
	serverSpec, err := definition.TranslateDefinition(context.TODO(), translated)
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

	return newResponse(record, server.StatusProvisioning, record.ExpiresAt)
}

func toOptions(options map[string]string) map[definition.Option]string {
	converted := make(map[definition.Option]string, len(options))
	for key, value := range options {
		converted[definition.Option(key)] = value
	}

	return converted
}

func (s *Service) shutdownExpiredServers() (expiredServers, error) {
	names, err := s.ShutdownExpired()
	if err != nil {
		return expiredServers{}, err
	}

	return expiredServers{Names: names}, nil
}

func (s *Service) ShutdownExpired() ([]string, error) {
	activeServerResources, err := s.clusterService.GetActiveServers()
	if err != nil {
		return nil, err
	}

	serversExpired := make([]string, 0)
	var failures []error
	for _, server := range activeServerResources.Items {
		remainingTime := getRemainingTime(server.Object)
		if remainingTime != remainingTime.Abs() {
			slog.Info("Server " + server.GetName() + " expired initiating shutdown")
			if err := s.clusterService.DeleteServer(server.GetName(), server.GetUID()); err != nil {
				slog.Error("Failed to delete expired server " + server.GetName() + ": " + err.Error())
				failures = append(failures, err)
				continue
			}
			serversExpired = append(serversExpired, server.GetName())
		} else {
			slog.Info("Server " + server.GetName() + " remaining time " + remainingTime.String())
		}
	}

	return serversExpired, errors.Join(failures...)
}
