package mgorm

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMysqlConfig(t *testing.T) {
	// user:password@tcp(host:port)/dbname?charset=utf8&parseTime=True&loc=Local
	dns := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	config := Config{
		DriverName: DriverNameMySql,
		Dns:        dns,
	}
	assert.Equal(t, dns, config.Dns)
}

func TestPostgresConfig(t *testing.T) {
	// host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai
	dns := "host=127.0.0.1 user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	config := Config{
		DriverName: DriverNamePostgres,
		Dns:        dns,
	}
	assert.Equal(t, dns, config.Dns)
}

func TestSqlite3Config(t *testing.T) {
	// file:filename?cache=shared&mode=ro
	dir, _ := os.Getwd()
	dns := dir + "/tmp/sqlite3.1.db"
	config := Config{
		DriverName: DriverNameSqlite3,
		Dns:        dns,
	}
	assert.Equal(t, dns, config.Dns)
}

func TestSqlServerConfig(t *testing.T) {
	// sqlserver://username:password@host/instance?param1=value&param2=value
	dns := "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
	config := Config{
		DriverName: DriverNameSqlServer,
		Dns:        dns,
	}
	assert.Equal(t, dns, config.Dns)
}
