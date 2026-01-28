package main

import (
	"seckill-project/common"     // 1. 引入数据库
	"seckill-project/controller" // 2. 引入控制器
	"seckill-project/utils"      // 3. 引入 utils 中的 JWT 中间件与工具

	"github.com/gin-gonic/gin"
)

func main() {
	common.InitDB()    // 先连数据库
	common.InitRedis() // 再连 Redis
	r := gin.Default()
	// 用户相关路由：注册与登录（不需要鉴权）
	r.POST("/register", controller.Register)
	r.POST("/login", controller.Login)
	// 商品相关路由
	r.POST("/product", controller.CreateProduct)
	r.GET("/product", utils.JWTAuthMiddleware(), controller.GetProductList)

	// 订单路由：下单前必须先登录，通过 JWT 中间件做鉴权
	r.POST("/order", utils.JWTAuthMiddleware(), controller.CreateOrder)
	r.Run(":8080")
}
