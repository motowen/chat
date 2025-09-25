package main

import (
	"fmt"
)

type User struct {
	UserID     string
	Supervisor string
	SectID     string
	DeptID     string
	DivisionID string
	FunctionID string
}

// cache: key = level:id, value = supervisorId
var supervisorCache = make(map[string]string)

// 統計 supervisor 出現次數，回傳票數最高的
func majoritySupervisor(users []User) string {
	count := make(map[string]int)
	for _, u := range users {
		if u.Supervisor != "" {
			count[u.Supervisor]++
		}
	}
	maxCount := -1
	winner := ""
	for k, v := range count {
		if v > maxCount {
			maxCount = v
			winner = k
		}
	}
	return winner
}

// 過濾符合某層級的 user
func filterUsers(users []User, level, id string) []User {
	var result []User
	for _, u := range users {
		switch level {
		case "sect":
			if u.SectID == id {
				result = append(result, u)
			}
		case "dept":
			if u.DeptID == id {
				result = append(result, u)
			}
		case "division":
			if u.DivisionID == id {
				result = append(result, u)
			}
		case "function":
			if u.FunctionID == id {
				result = append(result, u)
			}
		}
	}
	return result
}

// 遞迴尋找主管（含 cache）
func findSupervisor(users []User, level, id string) string {
	cacheKey := level + ":" + id
	if sup, ok := supervisorCache[cacheKey]; ok {
		return sup
	}

	group := filterUsers(users, level, id)
	if len(group) == 0 {
		supervisorCache[cacheKey] = ""
		return ""
	}

	// 判斷這層是否實際存在
	exists := false
	for _, u := range group {
		switch level {
		case "sect":
			if u.SectID != u.DeptID {
				exists = true
			}
		case "dept":
			if u.DeptID != u.DivisionID {
				exists = true
			}
		case "division":
			if u.DivisionID != u.FunctionID {
				exists = true
			}
		case "function":
			exists = true
		}
	}

	if exists {
		// 正常情況：統計這層 supervisor
		supID := majoritySupervisor(group)
		if supID == "" {
			supervisorCache[cacheKey] = ""
			return ""
		}
		// 找出 supervisor 的完整資料
		var supUser *User
		for _, u := range users {
			if u.UserID == supID {
				supUser = &u
				break
			}
		}
		if supUser == nil {
			supervisorCache[cacheKey] = supID
			return supID
		}

		// 根據層級往上遞迴
		switch level {
		case "sect":
			if supUser.DeptID != supUser.DivisionID {
				supID = findSupervisor(users, "dept", supUser.DeptID)
			}
		case "dept":
			if supUser.DivisionID != supUser.FunctionID {
				supID = findSupervisor(users, "division", supUser.DivisionID)
			}
		case "division":
			supID = findSupervisor(users, "function", supUser.FunctionID)
		case "function":
			// function 永遠是最上層
		}

		supervisorCache[cacheKey] = supID
		return supID
	}

	// 例外情況：這層不存在 → 直接跳到上層
	switch level {
	case "sect":
		return findSupervisor(users, "dept", group[0].DeptID)
	case "dept":
		return findSupervisor(users, "division", group[0].DivisionID)
	case "division":
		return findSupervisor(users, "function", group[0].FunctionID)
	case "function":
		supID := majoritySupervisor(group)
		supervisorCache[cacheKey] = supID
		return supID
	}

	supervisorCache[cacheKey] = ""
	return ""
}

func main() {
	users := []User{
		// 正常情況：有 sect → dept → division → function
		{UserID: "UA", Supervisor: "UB", SectID: "SA", DeptID: "DA", DivisionID: "DDA", FunctionID: "FA"},
		{UserID: "UC", Supervisor: "UB", SectID: "SA", DeptID: "DA", DivisionID: "DDA", FunctionID: "FA"},
		{UserID: "UB", Supervisor: "UD", SectID: "DA", DeptID: "DA", DivisionID: "DDA", FunctionID: "FA"},
		{UserID: "UD", Supervisor: "UF", SectID: "DDA", DeptID: "DDA", DivisionID: "DDA", FunctionID: "FA"},

		// 例外情況：沒有 sect/dept，直接跳 division
		{UserID: "UE", Supervisor: "UG", SectID: "DDB", DeptID: "DDB", DivisionID: "DDB", FunctionID: "FA"},
		{UserID: "UF", Supervisor: "UG", SectID: "DDB", DeptID: "DDB", DivisionID: "DDB", FunctionID: "FA"},
	}

	fmt.Println("Sect SA 主管:", findSupervisor(users, "sect", "SA"))
	fmt.Println("Dept DA 主管:", findSupervisor(users, "dept", "DA"))
	fmt.Println("Division DDA 主管:", findSupervisor(users, "division", "DDA"))
	fmt.Println("Function FA 主管:", findSupervisor(users, "function", "FA"))
	fmt.Println("Division DDB 主管 (例外):", findSupervisor(users, "division", "DDB"))

	// 測試 cache: 第二次呼叫同樣的查詢 → 會直接命中 cache
	fmt.Println("Cache 測試，Sect SA 再次查:", findSupervisor(users, "sect", "SA"))
}
