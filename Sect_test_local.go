package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupLocalMongo(t *testing.T) (*mongo.Client, context.Context, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	// 改成 docker-compose 的帳號密碼
	uri := "mongodb://root:example@localhost:27017"
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)

	cleanup := func() {
		client.Disconnect(ctx)
		cancel()
	}
	return client, ctx, cleanup
}

func TestGetSupervisors_LocalMongo(t *testing.T) {
	client, ctx, cleanup := setupLocalMongo(t)
	defer cleanup()

	coll := client.Database("testdb").Collection("users")
	_ = coll.Drop(ctx) // 確保乾淨環境

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

	out, _ := json.MarshalIndent(results, "", "  ")
	t.Logf("Results:\n%s", out)

	expected := []SupervisorResult{
		{SectId: "SA", SectSupervisor: "UA"},
		{SectId: "SB", SectSupervisor: "UA"},
		{SectId: "SC", SectSupervisor: "UD"},
		{DeptId: "DA", DeptSupervisor: "UB"},
		{DeptId: "DB", DeptSupervisor: "UB"},
	}
	require.ElementsMatch(t, expected, results)
}
