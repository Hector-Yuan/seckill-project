package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret 用来对 Token 进行签名和校验（对称密钥）
var jwtSecret = []byte("seckill-secret")

// Claims 自定义 Token 中存放的业务字段
// 这里除了用户 ID、用户名，还内嵌了标准的注册字段（过期时间等）
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 根据用户信息生成 JWT 字符串
func GenerateToken(userID uint, username string, role string) (string, error) {
	now := time.Now()
	// 构造负载（Payload）
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
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
