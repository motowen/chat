package main

import (
	"encoding/json"
	"fmt"
)

type User struct {
	SectId     string `json:"sectId"`
	DeptId     string `json:"deptId"`
	DivisionId string `json:"divisionId"`
	FunctionId string `json:"functionId"`
	UserId     string `json:"userId"`
	Supervisor string `json:"supervisor"`
}

// SectionSupervisor 輸出的 section 結構
type SectionSupervisor struct {
	SectId         string `json:"sectId"`
	SectSupervisor string `json:"sectSupervisor"`
}

// DeptSupervisor 輸出的 dept 結構
type DeptSupervisor struct {
	DeptId         string `json:"deptId"`
	DeptSupervisor string `json:"deptSupervisor"`
}

func main() {
	// 模擬從 Mongo 撈回來的資料
	users := []User{
		{"SA", "DA", "DDA", "FA", "UA", "UB"},
		{"SA", "DA", "DDA", "FA", "UZ", "UA"},
		{"SB", "DA", "DDA", "FA", "UW", "UA"},
		{"DA", "DA", "DDA", "FA", "UB", "UC"},
		{"SC", "DB", "DDA", "FA", "UD", "UB"},
		{"SC", "DB", "DDA", "FA", "UE", "UD"},
	}

	sectMap := make(map[string]string) // sectId -> supervisor
	deptMap := make(map[string]string) // deptId -> supervisor

	for _, u := range users {
		// 找 section 主管：只要 supervisor 存在，而且 supervisor 不是自己
		if u.Supervisor != "" {
			if _, exists := sectMap[u.SectId]; !exists {
				sectMap[u.SectId] = u.Supervisor
			}
		}

		// 找 dept 主管：條件可依實際需求調整
		// 這裡假設 userId == deptId 或有人 supervises 該 dept 的其他人
		if u.UserId == u.DeptId {
			deptMap[u.DeptId] = u.UserId
		} else if u.Supervisor != "" {
			if _, exists := deptMap[u.DeptId]; !exists {
				deptMap[u.DeptId] = u.Supervisor
			}
		}
	}

	// 整理輸出
	var result []interface{}

	for sect, sup := range sectMap {
		result = append(result, SectionSupervisor{SectId: sect, SectSupervisor: sup})
	}
	for dept, sup := range deptMap {
		result = append(result, DeptSupervisor{DeptId: dept, DeptSupervisor: sup})
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(out))
}
