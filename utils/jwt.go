package utils

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret 用来对 Token 进行签名和校验（对称密钥）
var jwtSecret = []byte("seckill-secret")

// Claims 自定义 Token 中存放的业务字段
// 这里除了用户 ID、用户名，还内嵌了标准的注册字段（过期时间等）
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 根据用户信息生成 JWT 字符串
func GenerateToken(userID uint, username string) (string, error) {
	now := time.Now()
	// 构造负载（Payload）
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // 过期时间：24 小时
			IssuedAt:  jwt.NewNumericDate(now),                     // 签发时间
			NotBefore: jwt.NewNumericDate(now),                     // 在此时间之前无效
		},
	}
	// 使用 HS256 算法生成 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析并校验传入的 Token 字符串
func ParseToken(tokenString string) (*Claims, error) {
	// 使用自定义的 Claims 结构进行解析
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 返回签名所用的密钥（和生成时保持一致）
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	// 校验 Token 是否有效，并且断言为我们自定义的 Claims 类型
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// JWTAuthMiddleware Gin 中间件：统一做登录态校验
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
		claims, err := ParseToken(tokenString)
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
