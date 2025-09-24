package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DebugChain 保留每個 user 的完整主管鏈
type DebugChain struct {
	UserId string   `bson:"userId" json:"userId"`
	Chain  []bson.M `bson:"chain" json:"chain"`
}

// SupervisorResult 表示 sect/dept 對應的主管
type SupervisorResult struct {
	SectId         string       `bson:"sectId" json:"sectId"`
	DeptId         string       `bson:"deptId" json:"deptId"`
	SectSupervisor string       `bson:"sectSupervisor" json:"sectSupervisor"`
	DeptSupervisor string       `bson:"deptSupervisor" json:"deptSupervisor"`
	DebugChains    []DebugChain `bson:"debugChains" json:"debugChains"`
}

// GetSupervisors 查詢每個 sect/dept 的主管
func GetSupervisors(ctx context.Context, coll *mongo.Collection) ([]SupervisorResult, error) {
	pipeline := mongo.Pipeline{
		{{"$graphLookup", bson.D{
			{"from", "users"},
			{"startWith", "$supervisor"},
			{"connectFromField", "supervisor"},
			{"connectToField", "userId"},
			{"as", "supervisors"},
			{"maxDepth", 10},
		}}},
		{{"$project", bson.D{
			{"sectId", 1},
			{"deptId", 1},
			{"userId", 1},
			{"supervisorsChain", "$supervisors"},
			{"sectSupervisor", bson.D{
				{"$ifNull", bson.A{
					bson.D{{"$arrayElemAt", bson.A{"$supervisors.userId", -1}}},
					"$supervisor",
				}},
			}},
			{"deptSupervisor", bson.D{
				{"$let", bson.D{
					{"vars", bson.D{
						{"deptSupervisors", bson.D{
							{"$filter", bson.D{
								{"input", "$supervisors"},
								{"cond", bson.D{{"$eq", bson.A{"$$this.sectId", "$$this.deptId"}}}},
							}},
						}},
					}},
					{"in", bson.D{
						{"$ifNull", bson.A{
							bson.D{{"$arrayElemAt", bson.A{"$$deptSupervisors.userId", 0}}},
							"$supervisor",
						}},
					}},
				}},
			}},
		}}},
		{{"$group", bson.D{
			{"_id", bson.D{{"sectId", "$sectId"}, {"deptId", "$deptId"}}},
			{"sectSupervisor", bson.D{{"$first", "$sectSupervisor"}}},
			{"deptSupervisor", bson.D{{"$first", "$deptSupervisor"}}},
			{"debugChains", bson.D{{"$push", bson.D{
				{"userId", "$userId"},
				{"chain", "$supervisorsChain"},
			}}}},
		}}},
		{{"$project", bson.D{
			{"_id", 0},
			{"sectId", "$_id.sectId"},
			{"deptId", "$_id.deptId"},
			{"sectSupervisor", 1},
			{"deptSupervisor", 1},
			{"debugChains", 1},
		}}},
	}

	cursor, err := coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []SupervisorResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	coll := client.Database("testdb").Collection("users")

	results, err := GetSupervisors(ctx, coll)
	if err != nil {
		panic(err)
	}

	for _, r := range results {
		fmt.Printf("Sect: %s, Dept: %s, SectSupervisor: %s, DeptSupervisor: %s\n",
			r.SectId, r.DeptId, r.SectSupervisor, r.DeptSupervisor)
		fmt.Printf("DebugChains: %+v\n", r.DebugChains)
		fmt.Println("-----")
	}
}
