package mgorm

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type DemoUser struct {
	ID       uint `gorm:"primaryKey;default:auto_random()"`
	Nickname string
}

func TestConnectionManager_Add(t *testing.T) {
	m := NewConnectionManager()
	m.Add("demo1", NewConnection(Config{
		DriverName: DriverNameMySql,
		Dns:        "root:root@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local",
	}))
	m.Add("demo2", NewConnection(Config{
		DriverName: DriverNameMySql,
		Dns:        "root:root@tcp(127.0.0.1:3306)/demo2?charset=utf8mb4&parseTime=True&loc=Local",
	}))

	db, err := m.Get("demo1").GetDB()
	if err != nil {
		assert.Error(t, err, "demo1: get db error")
	}
	db.AutoMigrate(&DemoUser{})
	createUser := &DemoUser{
		Nickname: fmt.Sprintf("demo1: TestConnection.GetDbWithMySQL.%s", time.Now().Format("2006-01-02 15:04:05")),
	}
	db.Create(createUser)

	db, err = m.Get("demo2").GetDB()
	if err != nil {
		assert.Error(t, err, "demo2: get db error")
	}
	db.AutoMigrate(&DemoUser{})
	createUser = &DemoUser{
		Nickname: fmt.Sprintf("demo2: TestConnection.GetDbWithMySQL.%s", time.Now().Format("2006-01-02 15:04:05")),
	}
	db.Create(createUser)
}

func TestDefaultConnectionManager(t *testing.T) {
	prefix := "default:"
	m := DefaultConnectionManager()
	m.Add("demo1", NewConnection(Config{
		DriverName: DriverNameMySql,
		Dns:        "root:root@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local",
	}))
	m.Add("demo2", NewConnection(Config{
		DriverName: DriverNameMySql,
		Dns:        "root:root@tcp(127.0.0.1:3306)/demo2?charset=utf8mb4&parseTime=True&loc=Local",
	}))

	db, err := m.Get("demo1").GetDB()
	if err != nil {
		assert.Error(t, err, prefix+"demo1: get db error")
	}
	db.AutoMigrate(&DemoUser{})
	createUser := &DemoUser{
		Nickname: fmt.Sprintf(prefix+"demo1: TestConnection.GetDbWithMySQL.%s", time.Now().Format("2006-01-02 15:04:05")),
	}
	db.Create(createUser)

	db, err = m.Get("demo2").GetDB()
	if err != nil {
		assert.Error(t, err, prefix+"demo2: get db error")
	}
	db.AutoMigrate(&DemoUser{})
	createUser = &DemoUser{
		Nickname: fmt.Sprintf(prefix+"demo2: TestConnection.GetDbWithMySQL.%s", time.Now().Format("2006-01-02 15:04:05")),
	}
	db.Create(createUser)
}
