package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const (
	baseURL      = "http://localhost:8080"
	numUsers     = 100 // 模拟的用户数量
	productStock = 10  // 商品库存
)

func main() {
	fmt.Printf("=== 开始准备压力测试数据 ===\n")
	fmt.Printf("模拟用户数: %d\n", numUsers)
	fmt.Printf("商品库存: %d\n", productStock)

	// 1. 准备用户和Token
	tokens := make([]string, numUsers)
	fmt.Println("正在注册并登录用户...")
	
	// 使用并发加速注册登录过程
	var prepareWg sync.WaitGroup
	prepareWg.Add(numUsers)
	for i := 0; i < numUsers; i++ {
		go func(index int) {
			defer prepareWg.Done()
			username := fmt.Sprintf("stress_user_%d_%d", time.Now().UnixNano(), index)
			password := "123456"
			tokens[index] = registerAndLogin(username, password)
		}(i)
	}
	prepareWg.Wait()
	fmt.Println("用户准备完成！")

	// 2. 创建一个测试商品
	// 假设第一个用户是管理员，或者用来创建商品
	adminToken := tokens[0]
	productID := createProduct(adminToken, "抗压测试手机", productStock)
	fmt.Printf("=== 商品创建成功 ID: %d, 库存: %d ===\n", productID, productStock)

	// 3. 开始压力测试
	fmt.Println("=== 3秒后开始并发抢购... ===")
	time.Sleep(3 * time.Second)

	var wg sync.WaitGroup
	var successCount int32
	var failCount int32
	var soldOutCount int32
	var errorCount int32

	startTime := time.Now()

	wg.Add(numUsers)
	for i := 0; i < numUsers; i++ {
		go func(token string, idx int) {
			defer wg.Done()
			status, _ := buy(token, productID)
			
			if status == 200 {
				atomic.AddInt32(&successCount, 1)
				fmt.Printf("[用户%d] 抢购成功\n", idx)
			} else if status == 400 {
				// 业务错误，可能是库存不足或重复购买
				// 这里我们可以简单归类为失败，或者细分
				atomic.AddInt32(&failCount, 1)
				// fmt.Printf("[用户%d] 抢购失败 (400)\n", idx)
			} else if status == 404 {
				atomic.AddInt32(&soldOutCount, 1) // 或者是商品不存在
			} else {
				atomic.AddInt32(&errorCount, 1)
				fmt.Printf("[用户%d] 请求异常 Status: %d\n", idx, status)
			}
		}(tokens[i], i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	fmt.Println("\n=== 测试结果统计 ===")
	fmt.Printf("耗时: %v\n", duration)
	fmt.Printf("总请求数: %d\n", numUsers)
	fmt.Printf("成功抢购: %d\n", successCount)
	fmt.Printf("抢购失败: %d (库存不足/重复下单)\n", failCount)
	fmt.Printf("其他错误: %d\n", errorCount)
	
	// 计算 QPS
	qps := float64(numUsers) / duration.Seconds()
	fmt.Printf("QPS: %.2f\n", qps)
	
	if int(successCount) > productStock {
		fmt.Printf("\n[严重错误] 超卖发生！库存: %d, 卖出: %d\n", productStock, successCount)
	} else {
		fmt.Printf("\n[正常] 未发生超卖。库存: %d, 卖出: %d\n", productStock, successCount)
	}
}

// --- Helper Functions ---

func registerAndLogin(username, password string) string {
	// 注册
	post(baseURL+"/register", map[string]interface{}{
		"username": username,
		"password": password,
	})
	// 登录
	resp, _ := post(baseURL+"/login", map[string]interface{}{
		"username": username,
		"password": password,
	})
	var res map[string]interface{}
	if err := json.Unmarshal(resp, &res); err != nil {
		fmt.Printf("Login response parse error for %s: %v\n", username, err)
		return ""
	}
	if token, ok := res["token"].(string); ok {
		return token
	}
	fmt.Printf("Login failed for %s: %v\n", username, res)
	return ""
}

func createProduct(token string, name string, stock int) int {
	resp, _ := postWithAuth(baseURL+"/product", token, map[string]interface{}{
		"name":  name,
		"stock": stock,
		"price": 100,
	})
	var res map[string]interface{}
	json.Unmarshal(resp, &res)
	if id, ok := res["id"].(float64); ok {
		return int(id)
	}
	return 0
}

func buy(token string, productID int) (int, []byte) {
	return postWithAuthResponse(baseURL+"/order", token, map[string]interface{}{
		"product_id": productID,
	})
}

func post(url string, data interface{}) ([]byte, error) {
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func postWithAuth(url, token string, data interface{}) ([]byte, error) {
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func postWithAuthResponse(url, token string, data interface{}) (int, []byte) {
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body
}
