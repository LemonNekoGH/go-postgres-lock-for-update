package main

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	// AppDB 数据库本库
	AppDB *gorm.DB
)

type testStruct struct {
	ID     string `gorm:"type:uuid"`
	Number int    // 要测试的数字
}

func (testStruct) TableName() string {
	return "public.lock_test"
}

// initDB 初始化数据库
func initDB() {
	// 初始化 Postgres 数据库
	dsn := "host=localhost user=postgres password=123456 dbname=postgres port=15432 sslmode=disable"
	var err error
	AppDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// 删除之前的表并创建新表
	err = AppDB.Migrator().DropTable(&testStruct{})
	if err != nil {
		panic(err)
	}
	err = AppDB.Migrator().CreateTable(&testStruct{})
	if err != nil {
		panic(err)
	}
}

func main() {
	initDB()

	// 写入 0
	id := uuid.NewString()
	err := AppDB.Create(&testStruct{ID: id, Number: 0}).Error
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			AppDB.Transaction(func(tx *gorm.DB) error {
				// 锁
				// err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id=?", id).Find(&testStruct{}).Error
				// if err != nil {
				// 	panic(err)
				// }
				// 可以尝试把注释去掉看看效果
				result := &testStruct{}
				err := tx.First(result).Error
				if err != nil {
					panic(err)
				}
				// 加一操作
				err = tx.Exec("update public.lock_test set number = ? where id = ?", result.Number+1, id).Error
				if err != nil {
					panic(err)
				}
				return nil
			})
		}()
	}

	wg.Wait()

	// 读取结果
	result := &testStruct{}
	err = AppDB.First(result).Error
	if err != nil {
		panic(err)
	}
	fmt.Printf("result: %d\n", result.Number)
}
