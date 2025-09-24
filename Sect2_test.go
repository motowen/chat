package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupMongoContainer(t *testing.T) (testcontainers.Container, *mongo.Client, context.Context, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:7.0", // 你可以換成需要的版本
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(20 * time.Second),
	}
	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// 取得 container 內部位址
	host, err := mongoC.Host(ctx)
	require.NoError(t, err)

	port, err := mongoC.MappedPort(ctx, "27017")
	require.NoError(t, err)

	uri := "mongodb://" + host + ":" + port.Port()

	// 建立 Mongo client
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	cleanup := func() {
		client.Disconnect(ctx)
		mongoC.Terminate(ctx)
	}

	return mongoC, client, ctx, cleanup
}

func TestGetSupervisors(t *testing.T) {
	_, client, ctx, cleanup := setupMongoContainer(t)
	defer cleanup()

	coll := client.Database("testdb").Collection("users")

	// 測試資料
	users := []interface{}{
		bson.M{"sectId": "SA", "deptId": "DA", "divisionId": "DDA", "functionId": "FA", "userId": "UA", "supervisor": "UB"},
		bson.M{"sectId": "SA", "deptId": "DA", "divisionId": "DDA", "functionId": "FA", "userId": "UZ", "supervisor": "UA"},
		bson.M{"sectId": "SB", "deptId": "DA", "divisionId": "DDA", "functionId": "FA", "userId": "UW", "supervisor": "UA"},
		bson.M{"sectId": "DA", "deptId": "DA", "divisionId": "DDA", "functionId": "FA", "userId": "UB", "supervisor": "UC"},
		bson.M{"sectId": "SC", "deptId": "DB", "divisionId": "DDA", "functionId": "FA", "userId": "UD", "supervisor": "UB"},
		bson.M{"sectId": "SC", "deptId": "DB", "divisionId": "DDA", "functionId": "FA", "userId": "UE", "supervisor": "UD"},
	}
	_, err := coll.InsertMany(ctx, users)
	require.NoError(t, err)

	// 執行查詢
	results, err := GetSupervisors(ctx, coll)
	require.NoError(t, err)

	// 輸出檢查
	out, _ := json.MarshalIndent(results, "", "  ")
	t.Logf("Results:\n%s", out)

	// 驗證
	expected := []SupervisorResult{
		{SectId: "SA", SectSupervisor: "UA"},
		{SectId: "SB", SectSupervisor: "UA"},
		{SectId: "SC", SectSupervisor: "UD"},
		{DeptId: "DA", DeptSupervisor: "UB"},
		{DeptId: "DB", DeptSupervisor: "UB"},
	}

	require.ElementsMatch(t, expected, results)
}
