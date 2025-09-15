package main

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) (*mongo.Client, *mongo.Database, context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("failed to connect mongo: %v", err)
	}

	db := client.Database("testdb_test")

	// 確保測試前是乾淨的
	db.Collection("employees").Drop(ctx)
	db.Collection("channels").Drop(ctx)

	return client, db, ctx
}

func TestAggregateWithMerge(t *testing.T) {
	client, db, ctx := setupTestDB(t)
	defer client.Disconnect(ctx)

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

	// 執行 aggregation pipeline
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
	defer cursor.Close(ctx)

	// 驗證 channels collection
	var results []bson.M
	if err := channels.Find(ctx, bson.M{}).All(&results); err != nil {
		t.Fatalf("find error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(results))
	}

	// 檢查其中一個 channel
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
