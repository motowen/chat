package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/sirupsen/logrus"
	"internal/app/router"
	"internal/pkg/cache"
	"internal/pkg/config"
	"internal/pkg/database"
	"internal/pkg/http/client"
	"test/container"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	redisContainer, setupRedisErr := container.SetupRedis(ctx)
	if setupRedisErr != nil {
		log.Fatalf("SetupRedis Fail, %s\n", setupRedisErr)
	}
	defer redisContainer.Terminate(ctx)

	mongoContainer, setupMongoErr := container.SetupMongo(ctx)
	if setupMongoErr != nil {
		log.Fatalf("SetupMongo Fail, %s\n", setupMongoErr)
	}
	defer mongoContainer.Terminate(ctx)

	Setup()
	httpmock.ActivateNonDefault(client.Get().GetClient())

	defer Close()
	go RunServer()

	r := m.Run()

	if r == 0 && testing.CoverMode() != "" {
		c := testing.Coverage() * 100
		l := 0.00
		fmt.Println("=================================================")
		fmt.Println("||               Coverage Report               ||")
		fmt.Println("=================================================")
		fmt.Printf("Cover mode: %s\n", testing.CoverMode())
		fmt.Printf("Coverage  : %.2f %% (Threshold: %.2f %%)\n\n", c, l)
		if c < l {
			fmt.Println("[Tests passed but coverage failed]")
			r = -1
		}
	}

	os.Exit(r)
}

func Setup() {
	var err error

	if err = config.Setup(); err != nil {
		log.Fatal(err)
	}

	if err = database.Setup(config.EnvVariable.DatabaseUri); err != nil {
		log.Fatal(err)
	}

	logLevel, err := logrus.ParseLevel(config.EnvVariable.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	logrus.SetLevel(logLevel)

	if err = cache.DefaultManager().Setup(cache.Config{
		Type:         config.EnvVariable.CacheType,
		EndpointList: config.EnvVariable.RedisEndpointList,
		Password:     config.EnvVariable.RedisPassword,
		Prefix:       config.EnvVariable.RedisPrefix,
	}); err != nil {
		log.Fatal(err)
	}

	if err = router.Setup(); err != nil {
		log.Fatal(err)
	}

	client.Setup()
}

func Close() {}

func RunServer() {
	s := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.EnvVariable.Port),
		Handler:      router.Router,
		ReadTimeout:  30 * time.Minute,
		WriteTimeout: 30 * time.Minute,
	}
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("%s\n", err)
	}
}
