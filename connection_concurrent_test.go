package mgorm

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"sync"
	"testing"
	"time"
)

type ConcurrentUser struct {
	ID       uint `gorm:"primaryKey;default:auto_random()"`
	Nickname string
}

func connectionCurrent(t *testing.T) *Connection {
	config := Config{
		DriverName: DriverNameMySql,
		Dns:        "root:root@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local",
	}
	con := NewConnection(config)
	err := con.Connect()
	if err != nil {
		assert.Error(t, err, "use mysql connect error")
	}
	return con
}

func connectionCurrentGetDB(t *testing.T) *gorm.DB {
	con := connectionCurrent(t)
	err := con.Connect()
	if err != nil {
		assert.Error(t, err, "use mysql connect error")
	}
	db, err := con.GetDB()
	if err != nil {
		assert.Error(t, err, "use mysql connect error")
	}
	// 创建表时有并发问题
	db.AutoMigrate(&ConcurrentUser{})
	return db
}

// TestConnectionConcurrent_Connect 测试并发连接
func TestConnectionConcurrent_Connect(t *testing.T) {
	con := connectionCurrent(t)
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := con.Connect()
			if err != nil {
				assert.Error(t, err, "use mysql connect error")
			}
		}()
	}
	wg.Wait()
}

// TestConnectionConcurrent_GetDB 并发测试
func TestConnectionConcurrent_GetDB(t *testing.T) {
	con := connectionCurrent(t)
	err := con.Connect()
	if err != nil {
		assert.Error(t, err, "use mysql connect error")
	}
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db, err := con.GetDB()
			if err != nil {
				assert.Error(t, err, "use mysql connect error")
			}
			// 创建表时有并发问题
			db.AutoMigrate(&ConcurrentUser{})
			/*
				2023/12/31 21:20:26 /Users/zhaoweijie/work/develop/project/2023/golang/mgorm/connection_concurrent_test.go:58 Error 1050 (42S01): Table 'concurrent_users' already exists
				[6.613ms] [rows:0] CREATE TABLE `concurrent_users` (`id` bigint unsigned AUTO_INCREMENT,`nickname` longtext,PRIMARY KEY (`id`))
			*/
			createUser := &ConcurrentUser{
				Nickname: fmt.Sprintf("demo1: TestConnection.GetDbWithMySQL.%s", time.Now().Format("2006-01-02 15:04:05")),
			}
			db.Create(createUser)
		}()
	}
	wg.Wait()
}
