package mgorm

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
)

// TestModel 测试用的数据模型
type TestModel struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"size:100"`
}

// ==================== NewManager 测试 ====================
// NewManager 返回 registry.Manager，用于管理多个资源组

// TestNewManager 测试 NewManager 函数创建多组数据库连接管理器
func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager() 返回 nil")
	}
}

// TestNewManager_AddGroupAndRegister 测试 Manager 添加组并注册连接
func TestNewManager_AddGroupAndRegister(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()

	// 添加资源组
	existed := manager.AddGroup("primary")
	if existed {
		t.Error("新组不应已存在")
	}

	// 获取组
	group, err := manager.Group("primary")
	if err != nil {
		t.Fatalf("Manager.Group() 失败: %v", err)
	}

	config := DBConfig{
		Name:            "测试数据库",
		Dialector:       sqlite.Open(":memory:"),
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: time.Hour,
	}

	// 注册资源
	isNew, err := group.Register(ctx, "db1", config)
	if err != nil {
		t.Fatalf("Group.Register() 失败: %v", err)
	}
	if !isNew {
		t.Error("应该是新资源")
	}

	// 获取数据库连接（惰性初始化）
	db, err := group.Get(ctx, "db1")
	if err != nil {
		t.Fatalf("Group.Get() 失败: %v", err)
	}
	if db == nil {
		t.Fatal("数据库连接不应为 nil")
	}

	// 验证连接可用
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层 sql.DB 失败: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}

	// 关闭管理器
	errs := manager.Close(ctx)
	if len(errs) != 0 {
		t.Fatalf("Manager.Close() 有错误: %v", errs)
	}
}

// TestNewManager_MultipleGroups 测试 Manager 管理多个组
func TestNewManager_MultipleGroups(t *testing.T) {
	ctx := context.Background()
	manager := NewManager()

	// 添加多个组
	manager.AddGroup("master")
	manager.AddGroup("slave1")
	manager.AddGroup("slave2")

	// 验证组列表
	groupNames := manager.ListGroupNames()
	if len(groupNames) != 3 {
		t.Errorf("期望 3 个组，实际 %d 个", len(groupNames))
	}

	// 在每个组注册数据库
	configs := map[string]DBConfig{
		"master": {
			Name:      "主数据库",
			Dialector: sqlite.Open(":memory:"),
		},
		"slave1": {
			Name:      "从库1",
			Dialector: sqlite.Open(":memory:"),
		},
		"slave2": {
			Name:      "从库2",
			Dialector: sqlite.Open(":memory:"),
		},
	}

	for groupName, cfg := range configs {
		group, err := manager.Group(groupName)
		if err != nil {
			t.Fatalf("获取组 %s 失败: %v", groupName, err)
		}
		_, err = group.Register(ctx, "db", cfg)
		if err != nil {
			t.Fatalf("注册 %s 的 db 失败: %v", groupName, err)
		}
	}

	// 验证所有连接可用
	for groupName := range configs {
		group, _ := manager.Group(groupName)
		db, err := group.Get(ctx, "db")
		if err != nil {
			t.Fatalf("获取 %s 的 db 失败: %v", groupName, err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("获取 %s 底层 sql.DB 失败: %v", groupName, err)
		}
		if err := sqlDB.Ping(); err != nil {
			t.Fatalf("Ping %s 失败: %v", groupName, err)
		}
	}

	// 清理
	errs := manager.Close(ctx)
	if len(errs) != 0 {
		t.Fatalf("Manager.Close() 有错误: %v", errs)
	}
}

// TestNewManager_MustGroup 测试 MustGroup 方法
func TestNewManager_MustGroup(t *testing.T) {
	manager := NewManager()
	manager.AddGroup("test")

	// 正常获取
	group := manager.MustGroup("test")
	if group == nil {
		t.Fatal("MustGroup() 返回 nil")
	}

	// 获取不存在的组应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGroup() 获取不存在的组应该 panic")
		}
	}()
	manager.MustGroup("nonexistent")
}

// TestNewManager_GetNonexistentGroup 测试获取不存在的组
func TestNewManager_GetNonexistentGroup(t *testing.T) {
	manager := NewManager()

	_, err := manager.Group("nonexistent")
	if err == nil {
		t.Error("获取不存在的组应返回错误")
	}
}

// ==================== New (NewGroup) 测试 ====================
// New 返回 registry.Group，用于单组资源管理

// TestNew 测试 New 函数创建单组数据库连接管理器
func TestNew(t *testing.T) {
	group := New()
	if group == nil {
		t.Fatal("New() 返回 nil")
	}
}

// TestNew_RegisterAndGet 测试 Group 的注册和获取连接功能
func TestNew_RegisterAndGet(t *testing.T) {
	ctx := context.Background()
	group := New()

	config := DBConfig{
		Name:      "主数据库",
		Dialector: sqlite.Open(":memory:"),
	}

	// 注册数据库连接
	isNew, err := group.Register(ctx, "primary", config)
	if err != nil {
		t.Fatalf("Group.Register() 失败: %v", err)
	}
	if !isNew {
		t.Error("应该是新资源")
	}

	// 通过名称获取连接
	db, err := group.Get(ctx, "primary")
	if err != nil {
		t.Fatalf("Group.Get() 失败: %v", err)
	}
	if db == nil {
		t.Fatal("数据库连接不应为 nil")
	}

	// 验证连接可用
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层 sql.DB 失败: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}

	// 清理
	errs := group.Close(ctx)
	if len(errs) != 0 {
		t.Fatalf("Group.Close() 有错误: %v", errs)
	}
}

// TestNew_MultipleConnections 测试 Group 管理多个数据库连接
func TestNew_MultipleConnections(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册多个数据库连接
	configs := map[string]DBConfig{
		"primary": {
			Name:      "主数据库",
			Dialector: sqlite.Open(":memory:"),
		},
		"replica1": {
			Name:      "从库1",
			Dialector: sqlite.Open(":memory:"),
		},
		"replica2": {
			Name:      "从库2",
			Dialector: sqlite.Open(":memory:"),
		},
	}

	for name, cfg := range configs {
		_, err := group.Register(ctx, name, cfg)
		if err != nil {
			t.Fatalf("注册 %s 失败: %v", name, err)
		}
	}

	// 验证所有连接都可以获取
	for name := range configs {
		db, err := group.Get(ctx, name)
		if err != nil {
			t.Errorf("获取 %s 连接失败: %v", name, err)
			continue
		}
		if db == nil {
			t.Errorf("%s 连接为 nil", name)
			continue
		}

		// 验证连接可用
		sqlDB, err := db.DB()
		if err != nil {
			t.Errorf("获取 %s 底层 sql.DB 失败: %v", name, err)
			continue
		}
		if err := sqlDB.Ping(); err != nil {
			t.Errorf("Ping %s 失败: %v", name, err)
		}
	}

	// 验证资源列表
	list := group.List()
	if len(list) != 3 {
		t.Errorf("期望 3 个资源，实际 %d 个", len(list))
	}

	// 清理
	errs := group.Close(ctx)
	if len(errs) != 0 {
		t.Fatalf("Group.Close() 有错误: %v", errs)
	}
}

// TestNew_GetUnregistered 测试获取未注册的连接
func TestNew_GetUnregistered(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 获取未注册的连接应返回错误
	_, err := group.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("获取未注册的连接应返回错误")
	}
}

// TestNew_MustGet 测试 MustGet 方法
func TestNew_MustGet(t *testing.T) {
	ctx := context.Background()
	group := New()

	config := DBConfig{
		Dialector: sqlite.Open(":memory:"),
	}
	group.Register(ctx, "db1", config)

	// 正常获取
	db := group.MustGet(ctx, "db1")
	if db == nil {
		t.Fatal("MustGet() 返回 nil")
	}

	// 获取不存在的资源应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet() 获取不存在的资源应该 panic")
		}
	}()
	group.MustGet(ctx, "nonexistent")
}

// TestNew_GormOperations 测试通过 Group 获取的连接执行 GORM 操作
func TestNew_GormOperations(t *testing.T) {
	ctx := context.Background()
	group := New()

	config := DBConfig{
		Name:      "GORM 分组测试",
		Dialector: sqlite.Open(":memory:"),
	}

	_, err := group.Register(ctx, "test_db", config)
	if err != nil {
		t.Fatalf("Group.Register() 失败: %v", err)
	}
	defer group.Close(ctx)

	db, err := group.Get(ctx, "test_db")
	if err != nil {
		t.Fatalf("Group.Get() 失败: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}

	// 创建记录
	record := TestModel{Name: "Group测试记录"}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("创建记录失败: %v", err)
	}
	if record.ID == 0 {
		t.Error("创建后记录 ID 不应为 0")
	}

	// 查询记录
	var found TestModel
	if err := db.First(&found, record.ID).Error; err != nil {
		t.Fatalf("查询记录失败: %v", err)
	}
	if found.Name != "Group测试记录" {
		t.Errorf("查询到的 Name = %q, 期望 %q", found.Name, "Group测试记录")
	}
}

// TestNew_ValidationError 测试 Group 处理无效配置
func TestNew_ValidationError(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 无效配置：没有 Dialector
	config := DBConfig{
		Name: "无效配置",
	}

	_, err := group.Register(ctx, "invalid", config)
	if err != nil {
		// Register 本身不会验证配置，配置会在 Get 时验证
		t.Logf("Register 返回错误（这是正常的如果实现了提前验证）: %v", err)
	}

	// Get 时会触发 opener，此时验证配置
	_, err = group.Get(ctx, "invalid")
	if err == nil {
		t.Error("Group.Get() 应返回错误（配置无效）")
	}
}

// TestNew_Unregister 测试注销连接
func TestNew_Unregister(t *testing.T) {
	ctx := context.Background()
	group := New()

	config := DBConfig{
		Dialector: sqlite.Open(":memory:"),
	}

	// 注册两个数据库连接
	group.Register(ctx, "db1", config)
	group.Register(ctx, "db2", config)

	// 获取连接（触发初始化）
	_, err := group.Get(ctx, "db1")
	if err != nil {
		t.Fatalf("获取 db1 失败: %v", err)
	}

	// 注销 db1
	if err := group.Unregister(ctx, "db1"); err != nil {
		t.Fatalf("Unregister() 失败: %v", err)
	}

	// db1 应该已注销
	_, err = group.Get(ctx, "db1")
	if err == nil {
		t.Error("注销后 db1 应返回错误")
	}

	// db2 应该仍然可用
	db2, err := group.Get(ctx, "db2")
	if err != nil {
		t.Fatalf("db2 应该仍可用: %v", err)
	}
	sqlDB, _ := db2.DB()
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("db2 Ping 失败: %v", err)
	}

	// 清理
	group.Close(ctx)
}

// TestNew_LazyInitialization 测试惰性初始化
func TestNew_LazyInitialization(t *testing.T) {
	ctx := context.Background()
	group := New()

	config := DBConfig{
		Dialector: sqlite.Open(":memory:"),
	}

	// 注册但不获取
	_, err := group.Register(ctx, "lazy_db", config)
	if err != nil {
		t.Fatalf("Register 失败: %v", err)
	}

	// 此时数据库连接尚未创建
	list := group.List()
	if len(list) != 1 {
		t.Errorf("期望 1 个资源，实际 %d 个", len(list))
	}

	// 第一次 Get 会触发初始化
	db1, err := group.Get(ctx, "lazy_db")
	if err != nil {
		t.Fatalf("第一次 Get 失败: %v", err)
	}

	// 第二次 Get 返回相同实例
	db2, err := group.Get(ctx, "lazy_db")
	if err != nil {
		t.Fatalf("第二次 Get 失败: %v", err)
	}

	if db1 != db2 {
		t.Error("多次 Get 应返回同一实例")
	}

	group.Close(ctx)
}

// ==================== opener 和 closer 函数测试 ====================

// TestOpener 测试 opener 函数
func TestOpener(t *testing.T) {
	ctx := context.Background()
	config := DBConfig{
		Dialector:       sqlite.Open(":memory:"),
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: time.Hour,
	}

	db, err := opener(ctx, config)
	if err != nil {
		t.Fatalf("opener() 失败: %v", err)
	}
	if db == nil {
		t.Fatal("opener() 返回 nil")
	}

	// 验证连接池配置
	sqlDB, _ := db.DB()
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 10 {
		t.Errorf("MaxOpenConnections = %d, 期望 10", stats.MaxOpenConnections)
	}

	// 清理
	sqlDB.Close()
}

// TestOpener_WithContext 测试 opener 函数的上下文支持
func TestOpener_WithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config := DBConfig{
		Dialector: sqlite.Open(":memory:"),
	}

	db, err := opener(ctx, config)
	if err != nil {
		t.Fatalf("opener() 失败: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// 验证连接可用
	sqlDB, _ := db.DB()
	if err := sqlDB.PingContext(ctx); err != nil {
		t.Errorf("PingContext 失败: %v", err)
	}
}

// TestOpener_ValidationError 测试 opener 函数处理无效配置
func TestOpener_ValidationError(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		config      DBConfig
		checkNoDSN  bool
		checkNoDial bool
	}{
		{
			name: "无 DSN 和 Dialector",
			config: DBConfig{
				Name: "无效配置",
			},
			checkNoDSN: true,
		},
		{
			name: "仅有 DSN 无 Dialector",
			config: DBConfig{
				DSN: "some_dsn_string",
			},
			checkNoDial: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := opener(ctx, tt.config)
			if err == nil {
				t.Error("opener() 应返回错误")
				if db != nil {
					sqlDB, _ := db.DB()
					sqlDB.Close()
				}
				return
			}
			if db != nil {
				t.Error("错误时数据库连接应为 nil")
			}
			if tt.checkNoDSN && !IsErrNoDSN(err) {
				t.Errorf("错误应为 NoDSN 类型，实际为: %v", err)
			}
			if tt.checkNoDial && !IsErrNoDialector(err) {
				t.Errorf("错误应为 NoDialector 类型，实际为: %v", err)
			}
		})
	}
}

// TestCloser 测试 closer 函数
func TestCloser(t *testing.T) {
	ctx := context.Background()
	config := DBConfig{
		Dialector: sqlite.Open(":memory:"),
	}

	db, err := opener(ctx, config)
	if err != nil {
		t.Fatalf("opener() 失败: %v", err)
	}

	// 测试关闭连接
	if err := closer(ctx, db); err != nil {
		t.Fatalf("closer() 失败: %v", err)
	}

	// 关闭后 Ping 应该失败
	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err == nil {
		t.Error("关闭后 Ping 应该失败")
	}
}

// TestCloser_NilDB 测试 closer 函数处理 nil 数据库
func TestCloser_NilDB(t *testing.T) {
	ctx := context.Background()

	// 传入 nil 应该返回 nil 而不是 panic
	if err := closer(ctx, nil); err != nil {
		t.Errorf("closer(nil) 应返回 nil，实际返回: %v", err)
	}
}

// ==================== 示例 ====================

// ExampleNewManager 展示 NewManager 的使用方法（多组管理）
func ExampleNewManager() {
	ctx := context.Background()

	// 创建多组数据库连接管理器
	manager := NewManager()

	// 添加主从组
	manager.AddGroup("master")
	manager.AddGroup("slave")

	// 获取主库组并注册连接
	masterGroup, _ := manager.Group("master")
	masterGroup.Register(ctx, "db1", DBConfig{
		Name:            "主库连接1",
		Dialector:       sqlite.Open(":memory:"),
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	})

	// 获取从库组并注册连接
	slaveGroup, _ := manager.Group("slave")
	slaveGroup.Register(ctx, "db1", DBConfig{
		Name:      "从库连接1",
		Dialector: sqlite.Open(":memory:"),
	})

	// 使用数据库连接
	masterDB, _ := masterGroup.Get(ctx, "db1")
	slaveDB, _ := slaveGroup.Get(ctx, "db1")

	_ = masterDB
	_ = slaveDB

	// 关闭所有连接
	manager.Close(ctx)
}

// ExampleNew 展示 New 的使用方法（单组管理）
func ExampleNew() {
	ctx := context.Background()

	// 创建单组数据库连接管理器
	group := New()

	// 注册主数据库
	group.Register(ctx, "primary", DBConfig{
		Name:      "主数据库",
		Dialector: sqlite.Open(":memory:"),
	})

	// 注册从数据库
	group.Register(ctx, "replica", DBConfig{
		Name:      "从数据库",
		Dialector: sqlite.Open(":memory:"),
	})

	// 获取并使用数据库连接
	primaryDB, _ := group.Get(ctx, "primary")
	replicaDB, _ := group.Get(ctx, "replica")

	_ = primaryDB
	_ = replicaDB

	// 关闭所有连接
	group.Close(ctx)
}
