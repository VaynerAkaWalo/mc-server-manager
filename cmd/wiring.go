package main

import (
	"context"
	"errors"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/bucket"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/cluster"
	"github.com/VaynerAkaWalo/mc-server-manager/internal/store"
	"os"
)

func connectStore(ctx context.Context) (*store.Store, error) {
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		return nil, errors.New("DB_URL is not set")
	}

	serverStore, err := store.Connect(ctx, dbURL)
	if err != nil {
		return nil, errors.New("Failed to connect to database: " + err.Error())
	}

	return serverStore, nil
}

func newClusterService() (*cluster.Service, error) {
	clientLoader := cluster.CreateClientLoader()

	clusterClient, err := clientLoader.Client()
	if err != nil {
		return nil, errors.New("Failed to create cluster client")
	}

	dynamicClient, err := clientLoader.DynamicClient()
	if err != nil {
		return nil, errors.New("Failed to create crd cluster client")
	}

	return cluster.CreateClusterService(*clusterClient, *dynamicClient), nil
}

func newSaveBucket() (*bucket.Client, error) {
	return bucket.New(bucket.Config{
		Endpoint:  os.Getenv("SAVES_ENDPOINT"),
		Bucket:    os.Getenv("SAVES_BUCKET"),
		Region:    os.Getenv("SAVES_REGION"),
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	})
}
