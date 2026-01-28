package controller

import (
	"seckill-project/common"
	"seckill-project/model"
	"seckill-project/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest 注册接口需要的入参结构体
type RegisterRequest struct {
	Username string `json:"username" binding:"required"` // 用户名必填
	Password string `json:"password" binding:"required"` // 密码必填
	Email    string `json:"email"`                       // 邮箱选填
}

// LoginRequest 登录接口需要的入参结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// Register 处理用户注册逻辑
func Register(c *gin.Context) {
	var req RegisterRequest
	// 解析 JSON 请求体到结构体中，并做基本的必填校验
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	// 对明文密码做 bcrypt 加密（哈希），存入数据库的永远是密文
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(500, gin.H{"error": "密码加密失败"})
		return
	}
	user := model.User{
		Username: req.Username,           // 用户名
		Password: string(hashedPassword), // 加密后的密码
		Email:    req.Email,              // 邮箱
	}
	// 写入数据库
	if err := common.DB.Create(&user).Error; err != nil {
		c.JSON(500, gin.H{"error": "注册失败"})
		return
	}
	c.JSON(200, gin.H{
		"message": "注册成功",
		"user_id": user.ID, // 把新用户的 ID 返回给前端
	})
}

// Login 处理用户登录逻辑
func Login(c *gin.Context) {
	var req LoginRequest
	// 解析 JSON 请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	// 根据用户名查询数据库中的用户
	var user model.User
	if err := common.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(400, gin.H{"error": "用户不存在"})
		return
	}
	// 使用 bcrypt 对比密码是否正确（明文 vs 加密后的密文）
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(400, gin.H{"error": "密码错误"})
		return
	}
	// 生成 JWT Token，后续请求要带在请求头中
	token, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": "生成令牌失败"})
		return
	}
	c.JSON(200, gin.H{
		"message":  "登录成功",
		"token":    token,
		"user_id":  user.ID,
		"username": user.Username,
	})
}
