package controller

import (
	"context"
	"encoding/json"
	"seckill-project/common" // 引入数据库包
	"seckill-project/model"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateProduct(c *gin.Context) {
	var p model.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(400, gin.H{"err": "输入格式不对"})
		return
	}
	result := common.DB.Create(&p)
	if result.Error != nil {
		// 如果有错，打印出来给前端看，你就知道原因了！
		c.JSON(500, gin.H{
			"err":    "入库失败",
			"reason": result.Error.Error(), // 这句话会告诉你具体的错误原因
		})
		return // 记得结束函数
	}
	c.JSON(200, gin.H{
		"msg": "成功啦",
		"id":  p.ID,
	})
}

func GetProductList(c *gin.Context) {
	// 1. 先试着从 Redis 缓存里拿数据
	ctx := context.Background()
	cacheKey := "product_list"

	val, err := common.RDB.Get(ctx, cacheKey).Result()
	if err == nil {
		// 缓存命中！直接反序列化并返回，不需要查数据库了
		var products []model.Product
		json.Unmarshal([]byte(val), &products)
		c.JSON(200, gin.H{
			"data":   products,
			"source": "cache", // 标记数据来源
		})
		return
	}

	// 2. 缓存没命中，去数据库查
	var products []model.Product
	// Find 不带条件就是查所有
	result := common.DB.Find(&products)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 3. 查到了，赶紧写入 Redis 缓存，有效期设为 10 秒（防止数据一直旧）
	jsonBytes, _ := json.Marshal(products)
	common.RDB.Set(ctx, cacheKey, jsonBytes, 10*time.Second)

	c.JSON(200, gin.H{
		"data":   products,
		"source": "database",
	})
}
