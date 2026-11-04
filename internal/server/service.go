package server

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/definition"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/mcping"
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

var (
	ErrNameInUse = errors.New("a server with this name already exists")
	ErrNotFound  = errors.New("server not found")

	ErrLoginRequired = errors.New("log in to use this tier")
	ErrSaveNotFound  = errors.New("save not found")
	ErrForeignOrigin = errors.New("invalid request origin")
	ErrSaveInUse     = errors.New("a server is still using this save, try again in a few minutes")
)

type serverStore interface {
	CreateServer(ctx context.Context, server store.NewServer) (store.ServerRecord, error)
	DeleteServer(ctx context.Context, id uuid.UUID) error
	CountServers(ctx context.Context, filter store.ServerFilter) (store.ServerCounts, error)
	ListServers(ctx context.Context, query store.ServerQuery) ([]store.ServerRecord, error)
	GetServer(ctx context.Context, id uuid.UUID) (store.ServerRecord, error)
	ListMonitoredServers(ctx context.Context) ([]store.ServerRecord, error)
	MarkRunning(ctx context.Context, id uuid.UUID) error
	MarkActive(ctx context.Context, id uuid.UUID) error
	MarkIdle(ctx context.Context, timeouts store.IdleTimeouts) error
	MarkDegraded(ctx context.Context, id uuid.UUID) error
	MarkStopped(ctx context.Context, id uuid.UUID) error
	MarkExpired(ctx context.Context) error
	GetSave(ctx context.Context, id uuid.UUID, owner string) (store.SaveRecord, error)
	LatestSaveDefinition(ctx context.Context, id uuid.UUID) (json.RawMessage, bool, error)
	ListReleasableSaveLocks(ctx context.Context) ([]store.ServerRecord, error)
	ReleaseSaveLock(ctx context.Context, id uuid.UUID) error
}

type clusterService interface {
	GetActiveServers() (*unstructured.UnstructuredList, error)
	DeployServerSpec(serverRequest serversv1alpha1.McServerSpec, serverID string) error
	DeleteServer(serverName string, uid types.UID) error
}

type Service struct {
	clusterService clusterService
	store          serverStore
	players        *playerCache
	ping           pingFunc
	versions       definition.VersionSource
	savesEnabled   bool
}

func CreateServerService(clusterService clusterService, store serverStore, versions definition.VersionSource) Service {
	return Service{
		clusterService: clusterService,
		store:          store,
		players:        newPlayerCache(),
		ping:           mcping.Ping,
		versions:       versions,
		savesEnabled:   true,
	}
}

func (s Service) WithoutSaves() Service {
	s.savesEnabled = false
	return s
}

func (s *Service) listServers(ctx context.Context, query store.ServerQuery) (server.Collection, error) {
	records, err := s.store.ListServers(ctx, query)
	if err != nil {
		return server.Collection{}, err
	}

	result := server.Collection{Items: make([]server.Response, 0, len(records))}
	if len(records) > query.Limit {
		records = records[:query.Limit]
		result.NextCursor = encodeCursor(query, records[len(records)-1])
	}
	for _, record := range records {
		response, err := s.respond(record)
		if err != nil {
			return server.Collection{}, err
		}
		result.Items = append(result.Items, response)
	}
	return result, nil
}

func (s *Service) getStats(ctx context.Context, filter store.ServerFilter) (stats, error) {
	counts, err := s.store.CountServers(ctx, filter)
	if err != nil {
		return stats{}, err
	}
	return stats{Online: counts.Online, Created: counts.Created}, nil
}

func (s *Service) provisionServer(ctx context.Context, owner string, provisionRequest server.Request) (server.Response, error) {
	serverDefinition := definition.ServerDefinition{
		Name:     provisionRequest.Name,
		Options:  toOptions(provisionRequest.Options),
		Tier:     definition.Tier(provisionRequest.Tier),
		Preset:   provisionRequest.Preset,
		Duration: definition.MaxServerLifetime,
		SaveId:   provisionRequest.SaveID,
	}

	translated := serverDefinition
	translated.Options = maps.Clone(serverDefinition.Options)
	serverSpec, err := definition.TranslateDefinition(ctx, translated, s.versions)
	if err != nil {
		return server.Response{}, err
	}

	serverDefinition.Options = maps.Clone(serverDefinition.Options)
	if serverDefinition.Options == nil {
		serverDefinition.Options = map[definition.Option]string{}
	}
	if version := serverSpec.Env["VERSION"]; version != "" {
		serverDefinition.Options[definition.VERSION] = version
	}

	save, err := s.checkSave(ctx, owner, serverDefinition)
	if err != nil {
		return server.Response{}, err
	}

	record, err := s.store.CreateServer(context.TODO(), store.NewServer{
		Name:       serverSpec.Name,
		Hostname:   serverAddress(serverSpec.Name),
		Definition: serverDefinition,
		ExpiresAt:  time.Now().Add(serverDefinition.Duration),
		Save:       save,
	})
	if errors.Is(err, store.ErrActiveNameTaken) {
		return server.Response{}, ErrNameInUse
	}
	if errors.Is(err, store.ErrSaveInUse) {
		return server.Response{}, ErrSaveInUse
	}
	if errors.Is(err, store.ErrSaveNotFound) {
		return server.Response{}, ErrSaveNotFound
	}
	if err != nil {
		return server.Response{}, err
	}

	err = s.clusterService.DeployServerSpec(*serverSpec, record.ID.String())
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

	return newResponse(record, server.StatusNew)
}

func (s *Service) checkSave(ctx context.Context, owner string, serverDefinition definition.ServerDefinition) (*store.SaveUse, error) {
	if owner == "" && definition.TierRequiresAccount(serverDefinition.Tier) {
		return nil, ErrLoginRequired
	}
	persistent := definition.TierPersistent(serverDefinition.Tier)
	if !s.savesEnabled && (serverDefinition.SaveId != "" || persistent) {
		return nil, definition.InvalidDefinitionError{Message: "saves are not available"}
	}
	if serverDefinition.SaveId == "" {
		if persistent {
			return nil, definition.InvalidDefinitionError{Message: "this tier needs a save"}
		}
		return nil, nil
	}
	if owner == "" {
		return nil, ErrLoginRequired
	}
	if !definition.TierPersistent(serverDefinition.Tier) {
		return nil, definition.InvalidDefinitionError{Message: "this tier cannot keep a save"}
	}

	id, err := uuid.Parse(serverDefinition.SaveId)
	if err != nil {
		return nil, definition.InvalidDefinitionError{Message: "the save id is not valid"}
	}
	_, err = s.store.GetSave(ctx, id, owner)
	if errors.Is(err, store.ErrSaveNotFound) {
		return nil, ErrSaveNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := s.checkNotDowngrade(ctx, id, serverDefinition); err != nil {
		return nil, err
	}

	return &store.SaveUse{ID: id, Owner: owner}, nil
}

func (s *Service) checkNotDowngrade(ctx context.Context, id uuid.UUID, requested definition.ServerDefinition) error {
	latest, found, err := s.store.LatestSaveDefinition(ctx, id)
	if err != nil || !found {
		return err
	}

	var played definition.ServerDefinition
	if err := json.Unmarshal(latest, &played); err != nil {
		return err
	}

	return definition.ValidateNotDowngrade(ctx, s.versions, played.Options[definition.VERSION], requested.Options[definition.VERSION])
}

func (s *Service) getServer(ctx context.Context, id uuid.UUID) (server.Response, error) {
	record, err := s.store.GetServer(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return server.Response{}, ErrNotFound
	}
	if err != nil {
		return server.Response{}, err
	}

	return s.respond(record)
}

func (s *Service) respond(record store.ServerRecord) (server.Response, error) {
	response, err := newResponse(record, server.Status(record.Status))
	if err != nil {
		return server.Response{}, err
	}
	if response.Status == server.StatusRunning {
		response.Players = s.players.get(record.ID)
	}

	return response, nil
}

func (s *Service) deleteServer(ctx context.Context, id uuid.UUID) error {
	record, err := s.store.GetServer(ctx, id)
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if !isActive(record.Status) {
		return nil
	}

	resources, err := s.clusterService.GetActiveServers()
	if err != nil {
		return err
	}
	for _, resource := range resources.Items {
		if resource.GetName() != record.Name {
			continue
		}
		if err := s.clusterService.DeleteServer(resource.GetName(), resource.GetUID()); err != nil {
			return err
		}
	}

	return s.store.MarkStopped(ctx, id)
}

func isActive(status string) bool {
	return status == string(server.StatusNew) || status == string(server.StatusRunning) || status == string(server.StatusDegraded) || status == string(server.StatusExpiring)
}

func toOptions(options map[string]string) map[definition.Option]string {
	converted := make(map[definition.Option]string, len(options))
	for key, value := range options {
		converted[definition.Option(key)] = value
	}

	return converted
}

func (s *Service) ShutdownExpired() ([]string, error) {
	if err := s.store.MarkExpired(context.TODO()); err != nil {
		return nil, err
	}

	activeServerResources, err := s.clusterService.GetActiveServers()
	if err != nil {
		return nil, err
	}

	activeRecords, err := s.store.ListMonitoredServers(context.TODO())
	if err != nil {
		return nil, err
	}
	active := make(map[string]bool, len(activeRecords))
	for _, record := range activeRecords {
		active[record.Name] = true
	}

	serversExpired := make([]string, 0)
	var failures []error
	for _, server := range activeServerResources.Items {
		if active[server.GetName()] {
			continue
		}

		slog.Info("Server " + server.GetName() + " expired initiating shutdown")
		if err := s.clusterService.DeleteServer(server.GetName(), server.GetUID()); err != nil {
			slog.Error("Failed to delete expired server " + server.GetName() + ": " + err.Error())
			failures = append(failures, err)
			continue
		}
		serversExpired = append(serversExpired, server.GetName())
	}

	return serversExpired, errors.Join(failures...)
}
