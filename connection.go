package mgorm

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"sync"
)

type Connection struct {
	db     *gorm.DB
	mutex  sync.Mutex
	config Config
}

// Connect 连接数据库
func (c *Connection) Connect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.db != nil {
		return nil
	}
	_, err := c.rawConnect()
	return err
}

// rawConnect 连接数据库
func (c *Connection) rawConnect() (*gorm.DB, error) {
	db, err := gorm.Open(Open(c.config), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	c.db = db
	return c.db, nil
}

// Disconnect 断开连接
func (c *Connection) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.db == nil {
		return nil
	}
	sqlDb, err := c.db.DB()
	if err != nil {
		return err
	}
	sqlDb.Close()
	c.db = nil
	return nil
}

func (c *Connection) GetDB() (*gorm.DB, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.db == nil {
		return c.rawConnect()
	}
	return c.db, nil
}

func Open(conf Config) gorm.Dialector {
	switch conf.DriverName {
	case DriverNameMySql:
		return mysql.Open(conf.Dns)
	case DriverNamePostgres:
		return postgres.Open(conf.Dns)
	case DriverNameSqlite3:
		return sqlite.Open(conf.Dns)
	case DriverNameSqlServer:
		return sqlserver.Open(conf.Dns)
	}
	return nil
}
