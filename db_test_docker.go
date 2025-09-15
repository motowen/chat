package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupMongoContainer(t *testing.T) (testcontainers.Container, *mongo.Client, context.Context) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mongo:6.0",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
	}
	mongoC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start mongo container: %v", err)
	}

	// 取得 host & port
	host, err := mongoC.Host(ctx)
	if err != nil {
		t.Fatalf("get host error: %v", err)
	}
	port, err := mongoC.MappedPort(ctx, "27017")
	if err != nil {
		t.Fatalf("get port error: %v", err)
	}

	uri := fmt.Sprintf("mongodb://%s:%s", host, port.Port())
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("failed to connect mongo: %v", err)
	}

	// 確保可 ping
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		t.Fatalf("ping mongo failed: %v", err)
	}

	// 測試結束自動清理
	t.Cleanup(func() {
		client.Disconnect(ctx)
		if err := mongoC.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	})

	return mongoC, client, ctx
}

func TestAggregateWithMerge(t *testing.T) {
	_, client, ctx := setupMongoContainer(t)

	db := client.Database("testdb")
	employees := db.Collection("employees")
	channels := db.Collection("channels")

	// 插入測試資料
	testDocs := []interface{}{
		bson.M{"account_id": "U1", "division_id": "A", "department_id": "D1", "section_id": "S1"},
		bson.M{"account_id": "U2", "division_id": "A", "department_id": "D1", "section_id": "S1"},
		bson.M{"account_id": "U3", "division_id": "B", "department_id": "D2", "section_id": "S2"},
	}
	_, err := employees.InsertMany(ctx, testDocs)
	if err != nil {
		t.Fatalf("insert error: %v", err)
	}

	// aggregation pipeline with $merge
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{"division_id", bson.D{{"$in", bson.A{"A", "B"}}}}}}},
		{{"$group", bson.D{
			{"_id", bson.D{
				{"division_id", "$division_id"},
				{"department_id", "$department_id"},
				{"section_id", "$section_id"},
			}},
			{"members", bson.D{{"$addToSet", "$account_id"}}},
		}}},
		{{"$project", bson.D{
			{"_id", 0},
			{"type", "section"},
			{"division_id", "$_id.division_id"},
			{"department_id", "$_id.department_id"},
			{"section_id", "$_id.section_id"},
			{"members", 1},
		}}},
		{{"$merge", bson.D{
			{"into", "channels"},
			{"whenMatched", "replace"},
			{"whenNotMatched", "insert"},
		}}},
	}

	cursor, err := employees.Aggregate(ctx, pipeline)
	if err != nil {
		t.Fatalf("aggregate error: %v", err)
	}
	cursor.Close(ctx)

	// 驗證 channels
	var results []bson.M
	cur, err := channels.Find(ctx, bson.M{})
	if err != nil {
		t.Fatalf("find error: %v", err)
	}
	if err := cur.All(ctx, &results); err != nil {
		t.Fatalf("cursor decode error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(results))
	}

	// 檢查 A/D1/S1 channel
	found := false
	for _, ch := range results {
		if ch["division_id"] == "A" && ch["department_id"] == "D1" && ch["section_id"] == "S1" {
			found = true
			members := ch["members"].(bson.A)
			if len(members) != 2 {
				t.Errorf("expected 2 members, got %v", members)
			}
		}
	}
	if !found {
		t.Errorf("expected to find channel for division A / department D1 / section S1")
	}
}
