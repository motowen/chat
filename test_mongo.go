package container

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/testcontainers/testcontainers-go"
)

type mongoContainer struct {
	testcontainers.Container
	URI string
}

func SetupMongo(ctx context.Context) (*mongoContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "mongo:3.6",
		Entrypoint:   nil,
		ExposedPorts: []string{"27017/tcp"},
		Name:         "mongo_test",
		AutoRemove:   true,
	}

	container, genericContainerErr := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if genericContainerErr != nil {
		return nil, genericContainerErr
	}

	mappedPort, mappedPortErr := container.MappedPort(ctx, "27017/tcp")
	if mappedPortErr != nil {
		return nil, mappedPortErr
	}

	hostIP, hostErr := container.Host(ctx)
	if hostErr != nil {
		return nil, hostErr
	}

	os.Setenv("DATABASE_URI", fmt.Sprintf("mongodb://%s:%s/adyen_gateway", hostIP, mappedPort.Port()))

	uri := fmt.Sprintf("mongodb://%s:%s/adyen_gateway", hostIP, mappedPort.Port())
	log.Println("mongo uri:", uri)

	return &mongoContainer{Container: container, URI: uri}, nil
}
