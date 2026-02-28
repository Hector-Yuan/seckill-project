package middleware

import (
	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 管理员权限校验
// 必须放在 JWTAuthMiddleware 之后使用，因为它依赖 Context 中的 role
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(401, gin.H{"error": "未获取到角色信息"})
			c.Abort()
			return
		}

		if roleStr, ok := role.(string); !ok || roleStr != "admin" {
			c.JSON(403, gin.H{"error": "权限不足，需要管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}
