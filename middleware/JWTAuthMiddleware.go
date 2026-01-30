package middleware

import (
	"seckill-project/utils"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中读取 Authorization 字段
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "未授权"})
			c.Abort()
			return
		}
		// 约定前缀为 "Bearer <token>"
		const prefix = "Bearer "
		if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
			c.JSON(401, gin.H{"error": "无效的认证头"})
			c.Abort()
			return
		}
		// 截取真正的 Token 字符串
		tokenString := authHeader[len(prefix):]
		// 调用工具函数解析 Token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}
		// 将用户信息放到 Gin 的 Context 中，后面的 Handler 可以直接获取
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		// 继续往下执行
		c.Next()
	}
}
