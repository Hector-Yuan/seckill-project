package common

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// RDB 全局 Redis 客户端
var RDB *redis.Client

// InitRedis 初始化 Redis 连接
func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 地址
		Password: "",               // 密码 (没有密码则留空)
		DB:       0,                // 默认 DB
	})

	// 测试连接
	ctx := context.Background()
	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		panic("连接 Redis 失败: " + err.Error())
	}
	fmt.Println("Redis 连接成功！")
}
