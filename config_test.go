package mgorm

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDBConfig_Validate_TableDriven 使用表驱动测试验证数据库配置校验
func TestDBConfig_Validate_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		config      DBConfig
		expectError bool
		checkNoDSN  bool // 是否检查错误类型为 NoDSN
		checkNoDial bool // 是否检查错误类型为 NoDialector
	}{
		{
			name: "有效配置：提供 Dialector",
			config: DBConfig{
				Dialector: sqlite.Open(":memory:"),
			},
			expectError: false,
		},
		{
			name: "有效配置：提供 Dialector 和其他参数",
			config: DBConfig{
				Name:            "测试数据库",
				DSN:             "ignored_when_dialector_set",
				DriverType:      "sqlite",
				Dialector:       sqlite.Open(":memory:"),
				MaxIdleConns:    10,
				MaxOpenConns:    100,
				ConnMaxLifetime: time.Hour,
			},
			expectError: false,
		},
		{
			name: "无效配置：Dialector 和 DSN 都为空",
			config: DBConfig{
				Name:       "空配置",
				DriverType: "mysql",
			},
			expectError: true,
			checkNoDSN:  true,
		},
		{
			name: "无效配置：只有 DSN 没有 Dialector",
			config: DBConfig{
				DSN:        "user:pass@tcp(localhost:3306)/dbname",
				DriverType: "mysql",
			},
			expectError: true,
			checkNoDial: true,
		},
		{
			name: "无效配置：空 DSN 字符串",
			config: DBConfig{
				DSN: "",
			},
			expectError: true,
			checkNoDSN:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Validate() 应返回错误，但返回 nil")
					return
				}
				if tt.checkNoDSN && !IsErrNoDSN(err) {
					t.Errorf("错误应为 NoDSN 类型，实际为: %v", err)
				}
				if tt.checkNoDial && !IsErrNoDialector(err) {
					t.Errorf("错误应为 NoDialector 类型，实际为: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() 不应返回错误，实际为: %v", err)
				}
			}
		})
	}
}

// TestOpenDB_Success 测试成功打开数据库连接
func TestOpenDB_Success(t *testing.T) {
	config := DBConfig{
		Name:            "测试内存数据库",
		Dialector:       sqlite.Open(":memory:"),
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: time.Minute,
	}

	db, err := openDB(config)
	if err != nil {
		t.Fatalf("openDB() 失败: %v", err)
	}

	if db == nil {
		t.Fatal("openDB() 返回的数据库连接不应为 nil")
	}

	// 验证连接可用
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层 sql.DB 失败: %v", err)
	}

	// 验证连接池配置
	stats := sqlDB.Stats()
	t.Logf("连接池统计: MaxOpenConnections=%d", stats.MaxOpenConnections)

	// 清理
	sqlDB.Close()
}

// TestOpenDB_ValidationError 测试配置验证失败的情况
func TestOpenDB_ValidationError(t *testing.T) {
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
			db, err := openDB(tt.config)

			if err == nil {
				t.Error("openDB() 应返回错误，但返回 nil")
				if db != nil {
					sqlDB, _ := db.DB()
					sqlDB.Close()
				}
				return
			}

			if db != nil {
				t.Error("openDB() 返回错误时，数据库连接应为 nil")
				sqlDB, _ := db.DB()
				sqlDB.Close()
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

// TestOpenDB_ConnectionPoolSettings 测试连接池参数配置
func TestOpenDB_ConnectionPoolSettings(t *testing.T) {
	tests := []struct {
		name            string
		maxIdleConns    int
		maxOpenConns    int
		connMaxLifetime time.Duration
	}{
		{
			name:            "默认值（零值）",
			maxIdleConns:    0,
			maxOpenConns:    0,
			connMaxLifetime: 0,
		},
		{
			name:            "自定义连接池参数",
			maxIdleConns:    5,
			maxOpenConns:    20,
			connMaxLifetime: 30 * time.Minute,
		},
		{
			name:            "仅设置最大空闲连接数",
			maxIdleConns:    10,
			maxOpenConns:    0,
			connMaxLifetime: 0,
		},
		{
			name:            "仅设置最大打开连接数",
			maxIdleConns:    0,
			maxOpenConns:    50,
			connMaxLifetime: 0,
		},
		{
			name:            "仅设置连接最大生存时间",
			maxIdleConns:    0,
			maxOpenConns:    0,
			connMaxLifetime: time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DBConfig{
				Dialector:       sqlite.Open(":memory:"),
				MaxIdleConns:    tt.maxIdleConns,
				MaxOpenConns:    tt.maxOpenConns,
				ConnMaxLifetime: tt.connMaxLifetime,
			}

			db, err := openDB(config)
			if err != nil {
				t.Fatalf("openDB() 失败: %v", err)
			}

			sqlDB, err := db.DB()
			if err != nil {
				t.Fatalf("获取底层 sql.DB 失败: %v", err)
			}
			defer sqlDB.Close()

			// 验证 MaxOpenConns 被正确设置
			if tt.maxOpenConns > 0 {
				stats := sqlDB.Stats()
				if stats.MaxOpenConnections != tt.maxOpenConns {
					t.Errorf("MaxOpenConnections = %d, 期望 %d", stats.MaxOpenConnections, tt.maxOpenConns)
				}
			}
		})
	}
}

// TestOpenDB_PingSuccess 测试数据库连接 Ping 功能
func TestOpenDB_PingSuccess(t *testing.T) {
	config := DBConfig{
		Name:      "Ping 测试数据库",
		Dialector: sqlite.Open(":memory:"),
	}

	db, err := openDB(config)
	if err != nil {
		t.Fatalf("openDB() 失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取底层 sql.DB 失败: %v", err)
	}
	defer sqlDB.Close()

	// 再次 Ping 验证连接仍然有效
	if err := sqlDB.Ping(); err != nil {
		t.Errorf("Ping 失败: %v", err)
	}
}

// TestDBConnection_Struct 测试 dbConnection 结构体
func TestDBConnection_Struct(t *testing.T) {
	config := DBConfig{
		Name:            "结构体测试",
		DSN:             "test_dsn",
		DriverType:      "sqlite",
		Dialector:       sqlite.Open(":memory:"),
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: time.Hour,
	}

	db, err := openDB(config)
	if err != nil {
		t.Fatalf("openDB() 失败: %v", err)
	}

	// 创建 dbConnection 实例
	conn := dbConnection{
		db:     db,
		config: config,
	}

	// 验证字段正确存储
	if conn.db == nil {
		t.Error("dbConnection.db 不应为 nil")
	}

	if conn.config.Name != "结构体测试" {
		t.Errorf("dbConnection.config.Name = %q, 期望 %q", conn.config.Name, "结构体测试")
	}

	if conn.config.MaxIdleConns != 5 {
		t.Errorf("dbConnection.config.MaxIdleConns = %d, 期望 %d", conn.config.MaxIdleConns, 5)
	}

	// 清理
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

// TestOpenDB_MultipleConnections 测试多次打开数据库连接
func TestOpenDB_MultipleConnections(t *testing.T) {
	config := DBConfig{
		Dialector: sqlite.Open(":memory:"),
	}

	// 打开多个连接
	connections := make([]*gorm.DB, 5)
	for i := 0; i < 5; i++ {
		db, err := openDB(config)
		if err != nil {
			t.Fatalf("第 %d 次 openDB() 失败: %v", i+1, err)
		}
		connections[i] = db
	}

	// 验证所有连接都是独立的
	for i := 0; i < 5; i++ {
		for j := i + 1; j < 5; j++ {
			if connections[i] == connections[j] {
				t.Errorf("连接 %d 和连接 %d 不应为同一实例", i, j)
			}
		}
	}

	// 清理所有连接
	for _, db := range connections {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}
}

// TestDBConfig_DefaultValues 测试 DBConfig 默认值
func TestDBConfig_DefaultValues(t *testing.T) {
	config := DBConfig{}

	// 验证默认值为零值
	if config.Name != "" {
		t.Errorf("默认 Name 应为空字符串，实际为 %q", config.Name)
	}
	if config.DSN != "" {
		t.Errorf("默认 DSN 应为空字符串，实际为 %q", config.DSN)
	}
	if config.DriverType != "" {
		t.Errorf("默认 DriverType 应为空字符串，实际为 %q", config.DriverType)
	}
	if config.MaxIdleConns != 0 {
		t.Errorf("默认 MaxIdleConns 应为 0，实际为 %d", config.MaxIdleConns)
	}
	if config.MaxOpenConns != 0 {
		t.Errorf("默认 MaxOpenConns 应为 0，实际为 %d", config.MaxOpenConns)
	}
	if config.ConnMaxLifetime != 0 {
		t.Errorf("默认 ConnMaxLifetime 应为 0，实际为 %v", config.ConnMaxLifetime)
	}
	if config.Dialector != nil {
		t.Error("默认 Dialector 应为 nil")
	}
}

// TestOpenDB_WithGormOperations 测试打开数据库后的 GORM 操作
func TestOpenDB_WithGormOperations(t *testing.T) {
	config := DBConfig{
		Name:      "GORM 操作测试",
		Dialector: sqlite.Open(":memory:"),
	}

	db, err := openDB(config)
	if err != nil {
		t.Fatalf("openDB() 失败: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// 定义测试模型
	type TestModel struct {
		ID   uint   `gorm:"primarykey"`
		Name string `gorm:"size:100"`
	}

	// 自动迁移
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate 失败: %v", err)
	}

	// 创建记录
	testRecord := TestModel{Name: "测试记录"}
	if err := db.Create(&testRecord).Error; err != nil {
		t.Fatalf("创建记录失败: %v", err)
	}

	if testRecord.ID == 0 {
		t.Error("创建后记录 ID 不应为 0")
	}

	// 查询记录
	var found TestModel
	if err := db.First(&found, testRecord.ID).Error; err != nil {
		t.Fatalf("查询记录失败: %v", err)
	}

	if found.Name != "测试记录" {
		t.Errorf("查询到的 Name = %q, 期望 %q", found.Name, "测试记录")
	}
}

// TestDBConfig_AutoDsn 测试 AutoDsn 方法
func TestDBConfig_AutoDsn(t *testing.T) {
	tests := []struct {
		name     string
		config   DBConfig
		expected string
	}{
		{
			name: "已有 DSN，直接返回",
			config: DBConfig{
				DSN: "existing_dsn_string",
			},
			expected: "existing_dsn_string",
		},
		{
			name: "MySQL 驱动，完整配置",
			config: DBConfig{
				DriverType: "mysql",
				Host:       "localhost",
				Port:       3306,
				User:       "root",
				Password:   "password",
				DBName:     "testdb",
				Charset:    "utf8mb4",
			},
			expected: "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "MySQL 驱动，使用默认字符集",
			config: DBConfig{
				DriverType: "mysql",
				Host:       "127.0.0.1",
				Port:       3306,
				User:       "admin",
				Password:   "secret",
				DBName:     "mydb",
			},
			expected: "admin:secret@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "PostgreSQL 驱动",
			config: DBConfig{
				DriverType: "postgres",
				Host:       "localhost",
				Port:       5432,
				User:       "postgres",
				Password:   "pgpass",
				DBName:     "postgresdb",
			},
			expected: "host=localhost port=5432 user=postgres password=pgpass dbname=postgresdb sslmode=disable",
		},
		{
			name: "SQLite 驱动，文件数据库",
			config: DBConfig{
				DriverType: "sqlite",
				DBName:     "/path/to/database.db",
			},
			expected: "/path/to/database.db",
		},
		{
			name: "SQLite 驱动，内存数据库",
			config: DBConfig{
				DriverType: "sqlite",
				DBName:     ":memory:",
			},
			expected: ":memory:",
		},
		{
			name: "SQL Server 驱动",
			config: DBConfig{
				DriverType: "sqlserver",
				Host:       "localhost",
				Port:       1433,
				User:       "sa",
				Password:   "mssqlpass",
				DBName:     "mssql_db",
			},
			expected: "sqlserver://sa:mssqlpass@localhost:1433?database=mssql_db",
		},
		{
			name: "未知驱动类型",
			config: DBConfig{
				DriverType: "unknown",
				Host:       "localhost",
				Port:       3306,
				User:       "user",
				Password:   "pass",
				DBName:     "db",
			},
			expected: "",
		},
		{
			name: "空驱动类型",
			config: DBConfig{
				DriverType: "",
				Host:       "localhost",
				Port:       3306,
				User:       "user",
				Password:   "pass",
				DBName:     "db",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.AutoDsn()
			if result != tt.expected {
				t.Errorf("AutoDsn() = %q, 期望 %q", result, tt.expected)
			}
		})
	}
}

// TestDBConfig_AutoDsn_MyCharsetModification 测试 AutoDsn 方法是否会修改原始配置的字符集
func TestDBConfig_AutoDsn_MyCharsetModification(t *testing.T) {
	config := DBConfig{
		DriverType: "mysql",
		Host:       "localhost",
		Port:       3306,
		User:       "root",
		Password:   "password",
		DBName:     "testdb",
		Charset:    "", // 故意留空
	}

	// 记录调用前的字符集
	originalCharset := config.Charset

	// 调用 AutoDsn
	dsn := config.AutoDsn()

	// 验证 DSN 包含默认字符集
	expectedDsn := "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != expectedDsn {
		t.Errorf("AutoDsn() = %q, 期望 %q", dsn, expectedDsn)
	}

	// 验证原始配置的字符集已被修改为默认值
	if config.Charset != "utf8mb4" {
		t.Errorf("调用 AutoDsn 后 Charset = %q, 期望 %q", config.Charset, "utf8mb4")
	}

	// 验证字符集确实发生了变化
	if config.Charset == originalCharset {
		t.Error("调用 AutoDsn 后 Charset 应该被修改")
	}
}

// TestDBConfig_AutoDsn_ExistingCharset 测试 AutoDsn 方法在已有字符集时的行为
func TestDBConfig_AutoDsn_ExistingCharset(t *testing.T) {
	config := DBConfig{
		DriverType: "mysql",
		Host:       "localhost",
		Port:       3306,
		User:       "root",
		Password:   "password",
		DBName:     "testdb",
		Charset:    "latin1", // 自定义字符集
	}

	dsn := config.AutoDsn()
	expectedDsn := "root:password@tcp(localhost:3306)/testdb?charset=latin1&parseTime=True&loc=Local"

	if dsn != expectedDsn {
		t.Errorf("AutoDsn() = %q, 期望 %q", dsn, expectedDsn)
	}

	// 验证字符集没有被修改
	if config.Charset != "latin1" {
		t.Errorf("Charset = %q, 期望保持为 latin1", config.Charset)
	}
}

// TestDBConfig_AutoDsn_EdgeCases 测试 AutoDsn 方法的边界情况
func TestDBConfig_AutoDsn_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		config   DBConfig
		expected string
	}{
		{
			name: "MySQL 配置缺少必要字段",
			config: DBConfig{
				DriverType: "mysql",
				Host:       "", // 缺少 Host
				Port:       3306,
				User:       "root",
				Password:   "password",
				DBName:     "testdb",
			},
			expected: "root:password@tcp(:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "MySQL 配置端口为零",
			config: DBConfig{
				DriverType: "mysql",
				Host:       "localhost",
				Port:       0, // 端口为零
				User:       "root",
				Password:   "password",
				DBName:     "testdb",
			},
			expected: "root:password@tcp(localhost:0)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "PostgreSQL 配置缺少字段",
			config: DBConfig{
				DriverType: "postgres",
				Host:       "", // 缺少 Host
				Port:       5432,
				User:       "postgres",
				Password:   "password",
				DBName:     "testdb",
			},
			expected: "host= port=5432 user=postgres password=password dbname=testdb sslmode=disable",
		},
		{
			name: "SQLite 空数据库名",
			config: DBConfig{
				DriverType: "sqlite",
				DBName:     "", // 空数据库名
			},
			expected: "", // 空字符串
		},
		{
			name: "SQL Server 配置缺少字段",
			config: DBConfig{
				DriverType: "sqlserver",
				Host:       "", // 缺少 Host
				Port:       1433,
				User:       "sa",
				Password:   "password",
				DBName:     "testdb",
			},
			expected: "sqlserver://sa:password@:1433?database=testdb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.AutoDsn()
			if result != tt.expected {
				t.Errorf("AutoDsn() = %q, 期望 %q", result, tt.expected)
			}
		})
	}
}

// BenchmarkDBConfig_AutoDsn 性能测试
func BenchmarkDBConfig_AutoDsn(b *testing.B) {
	configs := []DBConfig{
		{
			DSN: "existing_dsn",
		},
		{
			DriverType: "mysql",
			Host:       "localhost",
			Port:       3306,
			User:       "root",
			Password:   "password",
			DBName:     "testdb",
		},
		{
			DriverType: "postgres",
			Host:       "localhost",
			Port:       5432,
			User:       "postgres",
			Password:   "password",
			DBName:     "testdb",
		},
		{
			DriverType: "sqlite",
			DBName:     ":memory:",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := configs[i%len(configs)]
		_ = config.AutoDsn()
	}
}

// BenchmarkDBConfig_AutoDsn_MySQL MySQL DSN 生成性能测试
func BenchmarkDBConfig_AutoDsn_MySQL(b *testing.B) {
	config := DBConfig{
		DriverType: "mysql",
		Host:       "localhost",
		Port:       3306,
		User:       "root",
		Password:   "password",
		DBName:     "testdb",
		Charset:    "utf8mb4",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.AutoDsn()
	}
}
