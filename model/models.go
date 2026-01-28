package model // 👈 注意：这里必须叫 model，不能叫 main

import "gorm.io/gorm"

// User 用户表
type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(50);unique;not null" json:"username"`
	Password string `gorm:"type:varchar(100);not null" json:"-"`
	Email    string `gorm:"type:varchar(100)" json:"email"`
}

// Product 商品表
type Product struct {
	gorm.Model
	Name  string  `gorm:"type:varchar(100);not null" json:"name"`
	Title string  `gorm:"type:varchar(200)" json:"title"`
	Img   string  `gorm:"type:varchar(200)" json:"img"`
	Price float64 `gorm:"type:decimal(10,2)" json:"price"`
	Stock int     `gorm:"type:int" json:"stock"`
}

// Order 订单表
type Order struct {
	gorm.Model
	UserID    uint `json:"user_id" gorm:"uniqueIndex:idx_user_product"`    // 联合唯一索引
	ProductID uint `json:"product_id" gorm:"uniqueIndex:idx_user_product"` // 联合唯一索引
}
