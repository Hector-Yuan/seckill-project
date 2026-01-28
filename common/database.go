package common

import (
	"fmt"
	"seckill-project/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 全局数据库对象
var DB *gorm.DB

// InitDB 初始化数据库连接 (注意函数名首字母大写)
func InitDB() {
	dsn := "root:123456@tcp(127.0.0.1:3306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败: " + err.Error())
	}
	// 自动迁移
	DB.AutoMigrate(&model.User{}, &model.Product{}, &model.Order{})
	fmt.Println("数据库连接成功！")
}
