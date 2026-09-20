# 使用示例

```go
package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

// User 定义模型
type User struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"type:varchar(50);not nil;default:'';index"`
	Age       int            `gorm:"type:int;default:0"`
	Email     string         `gorm:"type:varchar(100);uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"` // 支持软删除
}

func InitDB() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 打印 SQL 日志
	})
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}
	// 自动迁移（自动创建/更新表结构）
	_ = db.AutoMigrate(&User{})
	return db
}


/* 
  增 Create 
*/
// 插入单条
user := User{Name:"zhangsan", Age: 18, Email: "zs@example.com"}
result := db.Create(&User) // 必须传指针，GORM会回填自增ID到user.ID
fmt.Println(user.ID, result.RowsAffected, result.Error)
// 批量插入
users := []User{
  {Name:"lisi", Age: 18, Email: "ls@example.com"},
  {Name:"wangwu", Age: 18, Email: "ww@example.com"},
}
db.Create(&User)

/* 
  删 Delete 
*/
// 逻辑删除，（因模型里有gorm.DeleteAt）
db.Delete(&User{}, 1) // 删除 id=1的纪录
db.Where("name = ?", "lisi").Delete(&User{})

// 物理删除
db.Unscoped().Delete(&User{}, 1)

/* 
  改 Update 
  Updates 传结构体时，会自动忽略零值(0, "", false)!
*/
// 如下代码只会更新 Name, 其他字段零值不会更新
db.Model(&User{}).Where("id=?", 2).Updates(User{Name:"newlisi", Age:0, Email: ""})

// 安全做法 1: 使用Map更新（支持零值）
db.Model(&User{}).Where("id=?", 2).Updates(map[string]any{"name": "newlisi", "age": 0, "email": ""})
// 安全做法 2: 使用Select强制更新字段）
db.Model(&User{}).Where("id=?", 2).Select("name", "age").Updates(User{Name:"newlisi", Age:0, Email: ""})

/* 
  查 Read 
*/
// 查询单条(First 如果查不到，会报gorm.ErrRecordNotFound错误)
var user User
err := db.First(&user, 1).Error // 主键查询

// 条件查询（Find查不到，只会返回空列表）
var users []User
db.Where("age > ? AND name LIKE ?", 18, "%张%").Find(&users)

// 分页与排序
db.Limit(10).Offset(20).Order("age DESC").Find(&users)

// 子查询
subQuery := db.Model(&User{}).Select("AVG(age)")
db.Where("age > ?", subQuery).Find(&users)

```

## 事务模式
```go
Update / Updates = 正常业务更新 + 触发 Hook + 自动更新时间
tx.Model(&User{}).
	Where("id=?", 1).
	Update("name", "newName")
// sql: UPDATE users SET name="newName", updated_at=now() where id = 1
// 多字段
tx.Model(&User{}).
	Where("id=?", 1).
	Updates(map[string]any{
		"name": "newName",
		"age": 18,
	})

UpdateColumn / UpdateColumns = 直接SQL更新 + 高性能 + 不触发 Hook + 不自动更新时间
tx.Model(&User{}).
	Where("id=?", 1).
	UpdateColumns(map[string]any{
		"name": "newName",
		"age": 18,
	})
// sql: UPDATE users SET name="newName", age=18 where id = 1

// 场景1: 计数器
tx.Model(&User{}).
	Where("id=?", 1).
	UpdateColumn(
		"likes",
		gorm.Expr("likes + ?", 1),
	)
// 场景2: 高并发余额
tx.Model(&User{}).
	Where("id=? and blance >= 100", 1).
	UpdateColumn(
		"balance"
		gorm.Expr("balance - ?", 100)	
	).
	Error
判断 rowsAffected == 1 为成功	

如果事务中需要额外判断逻辑 才加 for Update 来锁定行
tx.Clauses(
	clause.Locking{
		Strength:"UPDATE",
		// 1: join 表的时候用
		// Table: clause.Table{ 
    //     Name: "users",
    // },
		
		/*
		// 2: FOR UPDATE SKIP LOCKED
			已经被别人锁住的纪录直接跳过， 场景：Job Queue、多Worker并发消费
		*/ 
	},
).
Where("id = ?", 1).
First(&wallet)


err := db.Transaction(func(tx *gorm.DB) error {
    var wallet Wallet

    // 加行锁
    if err := tx.
        Clauses(clause.Locking{
            Strength: "UPDATE",
        }).
        Where("user_id = ?", uid).
        First(&wallet).Error; err != nil {
        return err
    }

    if wallet.Balance < amount {
        return errors.New("余额不足")
    }

    // 扣余额
    if err := tx.Model(&wallet).
        UpdateColumn(
            "balance",
            gorm.Expr("balance - ?", amount),
        ).Error; err != nil {
        return err
    }

    // 写流水
    if err := tx.Create(&WalletLog{
        UserID: uid,
        Amount: amount,
    }).Error; err != nil {
        return err
    }

    return nil
})
```