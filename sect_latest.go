package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User 資料結構
type User struct {
	UserId     string `bson:"userId"`
	SectId     string `bson:"sectId"`
	DeptId     string `bson:"deptId"`
	DivisionId string `bson:"divisionId"`
	FunctionId string `bson:"functionId"`
	Supervisor string `bson:"supervisor"`
}

// DepartmentSupervisorResult 報表結果
type DepartmentSupervisorResult struct {
	SectId           string
	SectSupervisor   string
	DeptId           string
	DeptSupervisor   string
	DivisionId       string
	DivisionSupervisor string
	FunctionId       string
	FunctionSupervisor string
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 連線 MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("yourDB").Collection("users")

	// 讀取所有 users
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var allUsers []User
	if err := cursor.All(ctx, &allUsers); err != nil {
		log.Fatal(err)
	}

	// 建立 userId -> User map
	userMap := make(map[string]*User)
	for i := range allUsers {
		u := &allUsers[i]
		userMap[u.UserId] = u
	}

	// 計算每個 user 的四層主管
	reportMap := make(map[string]DepartmentSupervisorResult)
	for _, u := range allUsers {
		sectSup, deptSup, divSup, funcSup := findSupervisors(u, userMap)

		// key 可以用四層部門組合
		key := fmt.Sprintf("%s|%s|%s|%s", u.SectId, u.DeptId, u.DivisionId, u.FunctionId)

		reportMap[key] = DepartmentSupervisorResult{
			SectId:            u.SectId,
			SectSupervisor:    sectSup,
			DeptId:            u.DeptId,
			DeptSupervisor:    deptSup,
			DivisionId:        u.DivisionId,
			DivisionSupervisor: divSup,
			FunctionId:        u.FunctionId,
			FunctionSupervisor: funcSup,
		}
	}

	// 輸出報表
	for _, r := range reportMap {
		fmt.Printf("%+v\n", r)
	}
}

// 找四層主管
func findSupervisors(user *User, userMap map[string]*User) (sectSup, deptSup, divSup, funcSup string) {
	sectSup, deptSup, divSup, funcSup = user.UserId, user.UserId, user.UserId, user.UserId // fallback = self
	visited := make(map[string]bool)

	current := userMap[user.Supervisor]
	for current != nil && !visited[current.UserId] {
		visited[current.UserId] = true

		if sectSup == user.UserId && current.SectId != user.SectId {
			sectSup = current.UserId
		}
		if deptSup == user.UserId && current.DeptId != user.DeptId {
			deptSup = current.UserId
		}
		if divSup == user.UserId && current.DivisionId != user.DivisionId {
			divSup = current.UserId
		}
		if funcSup == user.UserId && current.FunctionId != user.FunctionId {
			funcSup = current.UserId
		}

		current = userMap[current.Supervisor]
	}

	return
}
