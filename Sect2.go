package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 輸出資料結構 (sect 或 dept 主管，兩者共用一個 struct)
type SupervisorResult struct {
	SectId         string `bson:"sectId,omitempty" json:"sectId,omitempty"`
	SectSupervisor string `bson:"sectSupervisor,omitempty" json:"sectSupervisor,omitempty"`
	DeptId         string `bson:"deptId,omitempty" json:"deptId,omitempty"`
	DeptSupervisor string `bson:"deptSupervisor,omitempty" json:"deptSupervisor,omitempty"`
}

// 查詢 Sect/Dept Supervisor
func GetSupervisors(ctx context.Context, coll *mongo.Collection) ([]SupervisorResult, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$facet", Value: bson.D{
			{Key: "sect", Value: bson.A{
				bson.D{{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$sectId"},
					{Key: "sectSupervisor", Value: bson.D{{Key: "$first", Value: "$supervisor"}}},
				}}},
				bson.D{{Key: "$project", Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "sectId", Value: "$_id"},
					{Key: "sectSupervisor", Value: 1},
				}}},
			}},
			{Key: "dept", Value: bson.A{
				bson.D{{Key: "$group", Value: bson.D{
					{Key: "_id", Value: "$deptId"},
					{Key: "deptSupervisor", Value: bson.D{{Key: "$first", Value: bson.D{
						{Key: "$cond", Value: bson.A{
							bson.D{{Key: "$eq", Value: bson.A{"$userId", "$deptId"}}},
							"$userId",
							"$supervisor",
						}},
					}}}},
				}}},
				bson.D{{Key: "$project", Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "deptId", Value: "$_id"},
					{Key: "deptSupervisor", Value: 1},
				}}},
			}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "combined", Value: bson.D{{Key: "$concatArrays", Value: bson.A{"$sect", "$dept"}}}},
		}}},
		{{Key: "$unwind", Value: "$combined"}},
		{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$combined"}}}},
	}

	cur, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []SupervisorResult
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func main() {
	// 建立 Mongo Client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	coll := client.Database("testdb").Collection("users")

	results, err := GetSupervisors(ctx, coll)
	if err != nil {
		log.Fatal(err)
	}

	// 美化輸出 JSON
	out, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(out))
}
