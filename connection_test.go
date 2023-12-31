package mgorm

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type User struct {
	ID       uint `gorm:"primaryKey;default:auto_random()"`
	Nickname string
}

func TestConnection_ConnectWithMySQL(t *testing.T) {
	config := Config{
		DriverName: DriverNameMySql,
		Dns:        "root:root@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local",
	}
	con := NewConnection(config)
	err := con.Connect()
	if err != nil {
		assert.Error(t, err, "use mysql connect error")
	}
}

func TestConnection_DisconnectWithMySQL(t *testing.T) {
	config := Config{
		DriverName: DriverNameMySql,
		Dns:        "root:root@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local",
	}
	con := NewConnection(config)
	err := con.Connect()
	if err != nil {
		assert.Error(t, err, "use mysql connect error")
	}
	err = con.Disconnect()
	if err != nil {
		assert.Error(t, err, "use mysql disconnect error")
	}
}

func TestConnection_GetDbWithMySQL(t *testing.T) {
	config := Config{
		DriverName: DriverNameMySql,
		Dns:        "root:root@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local",
	}
	con := NewConnection(config)
	err := con.Connect()
	if err != nil {
		assert.Error(t, err, "use mysql connect error")
	}
	db, err := con.GetDB()
	if err != nil {
		assert.Error(t, err, "use mysql get db error")
	}
	db.AutoMigrate(&User{})
	createUser := &User{
		Nickname: fmt.Sprintf("TestConnection.GetDbWithMySQL.%s", time.Now().Format("2006-01-02 15:04:05")),
	}
	sqlDb := db.Create(createUser)
	assert.Equal(t, int64(1), sqlDb.RowsAffected, "TestSqlite3Connection.insertData.error", sqlDb.Error)
}
