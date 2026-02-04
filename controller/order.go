package controller

import (
	"fmt"
	"seckill-project/common"
	"seckill-project/model"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

// CreateOrder
// 职责：处理“秒杀下单”请求，保证在高并发场景下不超卖、不重复购入，并且所有相关数据库写入具备原子性
// 关键设计：
// - 从 JWT 中间件注入的上下文中拿到用户身份，确保只有已登录用户才能下单
// - 使用单个数据库事务串联“查重 → 加锁读取库存 → 扣减库存 → 创建订单”，实现要么全部成功、要么全部回滚
// - 通过悲观锁（SELECT ... FOR UPDATE）在行级别锁住目标商品，避免并发下的“读-改-写”互相踩踏导致库存回退或超卖
// - 先查重后扣减，结合表结构中的联合唯一索引（user_id + product_id），防止同一用户对同一商品重复下单
// - 所有读写均使用同一个事务句柄 tx，保证隔离性与可见性一致，避免“读到旧值/新值不一致”的问题
func CreateOrder(c *gin.Context) {
	// 1) 解析入参：前端只需传 product_id
	// 说明：ShouldBindJSON 会从请求体读取 JSON，根据结构体字段上的 json tag 进行绑定；
	//      如果 JSON 格式错误、类型不匹配或缺少字段，会返回错误并直接终止。
	type OrderRequest struct {
		ProductID uint `json:"product_id"`
	}
	var req OrderRequest
	// 把前端传来的 JSON 请求体解析到 req 结构体中；失败则返回 400
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 2) 获取当前登录用户：JWT 中间件已在 Context 注入 userID/username
	// 说明：中间件通过 ParseToken 校验后，将用户信息写入 Context；
	//      此处只需要安全地读取并做类型断言即可。
	userIDValue, ok := c.Get("userID")
	if !ok {
		c.JSON(401, gin.H{"error": "未授权"})
		return
	}
	// 断言为 uint，与 Claims 中的类型保持一致；失败视为服务端状态异常
	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(500, gin.H{"error": "用户信息错误"})
		return
	}
	//Redis锁
	lockKey := fmt.Sprintf("order_lock:%d:%d", userID, req.ProductID)
	lockkeySuccess, err := common.RDB.SetNX(c.Request.Context(), lockKey, "1", 5*time.Second).Result()
	if err != nil {
		// Redis 连不上或其他网络错误
		c.JSON(500, gin.H{"error": "系统繁忙，请稍后再试"})
		return
	}
	if !lockkeySuccess {
		// 锁已存在，说明有人（或者同一个用户）正在操作
		c.JSON(400, gin.H{"error": "操作太频繁，请稍后再试"})
		return
	}
	// 抢锁成功后，无论后续业务成功还是失败，都最好在函数结束时释放锁
	// 这样可以避免：万一业务逻辑只执行了 10ms，锁却占用了 5s，
	// 虽然对“防刷”来说 5s 也没问题，但作为标准分布式锁，用完即删是好习惯。
	defer common.RDB.Del(c.Request.Context(), lockKey)

	// 3) 开启事务：后续所有数据库操作统一使用 tx
	// 说明：事务保证 ACID 特性：
	//  - 原子性：其中任一步失败则整体回滚
	//  - 一致性：库存与订单始终保持逻辑一致
	//  - 隔离性：并发事务之间互不干扰（结合锁实现）
	//  - 持久性：提交后数据可靠落库
	tx := common.DB.Begin()

	// 4) 查重：同一用户对同一商品仅允许一单
	// 说明：此处的业务校验与数据库层联合唯一索引相互配合，
	//      先在事务内做快速判定，能更友好地向用户返回提示；
	//      极端并发表的唯一索引仍是兜底保障。
	var count int64
	tx.Model(&model.Order{}).Where("user_id = ? AND product_id = ?", userID, req.ProductID).Count(&count)
	if count > 0 {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "你已经抢过啦，每人限购一件！"})
		return
	}

	// 5) 加锁读取商品库存：防止并发下“同时读、同时减”造成超卖
	// 说明：SELECT ... FOR UPDATE 会给目标行加排他锁（InnoDB 行锁），
	//      直到事务提交/回滚才释放；其他事务在此期间读取同一行的更新会被阻塞，
	//      从而形成串行化的扣减过程。
	var product model.Product
	// 悲观锁用法：通过 gorm 的 query_option 注入 FOR UPDATE，确保读取到的是锁定版本
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, req.ProductID).Error; err != nil {
		// 商品不存在或查询异常：直接回滚并返回
		tx.Rollback()
		c.JSON(404, gin.H{"error": "商品不存在"})
		return
	}

	// 6) 校验库存是否可售：库存必须大于 0
	// 说明：由于当前行已加锁，这个判断在本事务内具备并发安全性。
	if product.Stock <= 0 {
		tx.Rollback()
		c.JSON(400, gin.H{"error": "手慢无，卖光啦！"})
		return
	}

	// 7) 扣减库存：在同一事务内对锁定行执行更新
	// 说明：使用 tx.Model(&product).Update(...) 保证对同一行的原子更新；
	//      若更新失败（例如数据库错误或锁冲突），则整体回滚。
	if err := tx.Model(&product).Update("stock", product.Stock-1).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "扣减库存失败"})
		return
	}

	// 8) 创建订单：与库存扣减同属一个事务
	// 说明：若此处因唯一索引或其他原因失败，则回滚前面的扣减，保持业务一致。
	order := model.Order{
		UserID:    userID,
		ProductID: product.ID,
	}
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "创建订单失败"})
		return
	}

	// 9) 提交事务：库存扣减与订单创建同时生效
	// 说明：提交成功后才会释放行锁，并对外可见；失败则由数据库返回错误。
	tx.Commit()

	c.JSON(200, gin.H{
		"message":  "抢购成功！",
		"order_id": order.ID,
	})
}
