package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// baseURL 是被压测的后端服务地址
// 在当前项目中就是你 main.go 起起来的 Gin 服务
const baseURL = "http://localhost:8080"

// main 是并发测试的入口函数：
// 1）注册两个用户并登录拿到各自的 token
// 2）创建一个库存只有 1 的商品
// 3）用两个用户并发去抢同一个商品，观察抢购结果
func main() {
	// 1. 先注册 2 个用户，并拿到各自的 token
	token1 := registerAndLogin("test_u1", "123456")
	token2 := registerAndLogin("test_u2", "123456")

	// 2. 创建一个商品，库存设为 1（只允许成功一单）
	productID := createProduct(token1, "秒杀手机", 1) // 假设 token1 有权限创建商品
	fmt.Printf("=== 商品创建成功 ID: %d, 库存: 1 ===\n", productID)

	// 3. 并发抢购：两个用户同时对同一个商品下单
	var wg sync.WaitGroup
	var mu sync.Mutex // 互斥锁：保护下面的计数器
	var successCount int
	var failCount int

	wg.Add(2)

	fmt.Println("=== 开始并发抢购 ===")
	go func() {
		defer wg.Done()
		success := buy(token1, productID, "用户1")

		mu.Lock() // 加锁
		if success {
			successCount++
		} else {
			failCount++
		}
		mu.Unlock() // 解锁
	}()

	go func() {
		defer wg.Done()
		success := buy(token2, productID, "用户2")

		mu.Lock() // 加锁
		if success {
			successCount++
		} else {
			failCount++
		}
		mu.Unlock() // 解锁
	}()

	wg.Wait()
	fmt.Printf("=== 测试结束 | 成功: %d | 失败: %d ===\n", successCount, failCount)
}

// registerAndLogin 负责帮你：
// 1）调用 /register 注册用户
// 2）调用 /login 登录并返回登录成功后的 token
func registerAndLogin(username, password string) string {
	// 注册：这里不关心结果细节，只要调用过去即可
	post(baseURL+"/register", map[string]interface{}{
		"username": username,
		"password": password,
	})
	// 登录：从返回值中解析出 token
	resp := post(baseURL+"/login", map[string]interface{}{
		"username": username,
		"password": password,
	})
	var res map[string]interface{}
	json.Unmarshal(resp, &res)
	return res["token"].(string)
}

// createProduct 调用 /product 接口创建商品，并返回商品 ID
func createProduct(token string, name string, stock int) int {
	resp := postWithAuth(baseURL+"/product", token, map[string]interface{}{
		"name":  name,
		"stock": stock,
		"price": 100,
	})
	var res map[string]interface{}
	json.Unmarshal(resp, &res)
	return int(res["id"].(float64))
}

// buy 调用 /order 接口，模拟某个用户对指定商品发起抢购请求
// 返回值：true 表示抢购成功，false 表示失败
func buy(token string, productID int, userLabel string) bool {
	resp := postWithAuth(baseURL+"/order", token, map[string]interface{}{
		"product_id": productID,
	})

	respStr := string(resp)
	fmt.Printf("[%s] 抢购结果: %s\n", userLabel, respStr)

	// 简单判断返回内容中是否包含 "id" 字段来判定是否成功
	// (实际项目中应该解析 JSON 判断 status code)
	return bytes.Contains(resp, []byte("\"id\""))
}

// post 是一个简单的工具函数，用于发送无需鉴权的 POST 请求（例如注册、登录）
func post(url string, data interface{}) []byte {
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body
}

// postWithAuth 用于发送带 JWT 鉴权头的 POST 请求（例如创建商品、下单）
func postWithAuth(url, token string, data interface{}) []byte {
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body
}
