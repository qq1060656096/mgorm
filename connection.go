package mgorm

import (
	"gorm.io/gorm"
	"sync"
)

type Connection struct {
	db    *gorm.DB
	mutex sync.Mutex
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

func (c *Connection) rawConnect() (*gorm.DB, error) {
	db, err := gorm.Open(NewDialector(conn.conf), &gorm.Config{})
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
