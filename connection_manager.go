package mgorm

import (
	"fmt"
	"sync"
)

// 默认连接管理器
var defaultConnectionManager = NewConnectionManager()

// ConnectionManager 连接管理器
type ConnectionManager struct {
	conMap sync.Map
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{}
}

// Add 添加连接
func (m *ConnectionManager) Add(name string, con *Connection) {
	m.conMap.LoadOrStore(name, con)
}

// Delete 删除连接
func (m *ConnectionManager) Delete(name string) {
	m.conMap.Delete(name)
}

// Get 获取连接
func (m *ConnectionManager) Get(name string) *Connection {
	if c, ok := m.conMap.Load(name); ok {
		fmt.Println(">>-----", ok, c.(*Connection))
		return c.(*Connection)
	}
	return nil
}

// Exist 判断连接是否存在
func (m *ConnectionManager) Exist(name string) bool {
	_, ok := m.conMap.Load(name)
	return ok
}

// DefaultConnectionManager 默认连接管理器
func DefaultConnectionManager() *ConnectionManager {
	return defaultConnectionManager
}
