package main

import (
	"seckill-project/common"     // 1. 引入数据库
	"seckill-project/controller" // 2. 引入控制器
	"seckill-project/middleware" // 3. 引入 middleware 中的 JWT 中间件

	"github.com/gin-gonic/gin"
)

func main() {
	common.InitDB()    // 先连数据库
	common.InitRedis() // 再连 Redis
	r := gin.Default()
	publicGroup := r.Group("/")
	{
		publicGroup.POST("/register", controller.Register)
		publicGroup.POST("/login", controller.Login)
	}
	privategroup := r.Group("/")
	privategroup.Use(middleware.JWTAuthMiddleware())
	{
		privategroup.POST("/product", controller.CreateProduct)
		privategroup.GET("/product", controller.GetProductList)
		privategroup.POST("/order", controller.CreateOrder)
	}

	// 订单路由：下单前必须先登录，通过 JWT 中间件做鉴权
	r.Run(":8080")
}
