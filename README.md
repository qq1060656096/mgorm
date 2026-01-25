# mgorm

基于 [GORM](https://gorm.io/) 的数据库连接管理库，提供连接池配置、多数据库实例管理等功能。

本库基于 [bizutil/registry](https://github.com/qq1060656096/bizutil/tree/main/registry) 包实现资源管理功能。

## 功能特性

- ✨ 基于 GORM 的数据库连接管理
- 🔄 连接池配置（最大空闲连接数、最大打开连接数、连接最大存活时间）
- 📦 多数据库实例管理（单组管理 / 多组管理）
- ⚡ 惰性初始化（首次获取时创建连接）
- 🔒 线程安全

## 安装

```bash
go get github.com/qq1060656096/mgorm
```

**使用 MySQL 时需要额外安装驱动：**

```bash
go get gorm.io/driver/mysql
```

## 快速开始

### 基础用法（单组管理）

适用于管理多个命名的数据库连接：

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/qq1060656096/mgorm"
    "gorm.io/driver/mysql"
)

func main() {
    ctx := context.Background()

    // 创建数据库连接组
    group := mgorm.New()

    // 注册主数据库
    _, err := group.Register(ctx, "primary", mgorm.DBConfig{
        Name:            "主数据库",
        DSN:             "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        Dialector:       mysql.Open("user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"),
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 注册从数据库
    _, err = group.Register(ctx, "replica", mgorm.DBConfig{
        Name:            "从数据库",
        DSN:             "user:password@tcp(127.0.0.1:3307)/dbname?charset=utf8mb4&parseTime=True&loc=Local",
        Dialector:       mysql.Open("user:password@tcp(127.0.0.1:3307)/dbname?charset=utf8mb4&parseTime=True&loc=Local"),
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 获取数据库连接
    primaryDB, err := group.Get(ctx, "primary")
    if err != nil {
        log.Fatal(err)
    }

    replicaDB, err := group.Get(ctx, "replica")
    if err != nil {
        log.Fatal(err)
    }

    // 使用数据库连接
    var result []map[string]interface{}
    primaryDB.Raw("SELECT 1").Scan(&result)
    replicaDB.Raw("SELECT 1").Scan(&result)

    // 程序退出时关闭所有连接
    defer group.Close(ctx)
}
```

### 多组管理

适用于需要管理多个数据库组的场景（如主从分离、多租户等）：

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/qq1060656096/mgorm"
    "gorm.io/driver/mysql"
)

func main() {
    ctx := context.Background()

    // 创建多组数据库连接管理器
    manager := mgorm.NewManager()

    // 添加主库组和从库组
    manager.AddGroup("master")
    manager.AddGroup("slave")

    // 获取主库组并注册连接
    masterGroup, err := manager.Group("master")
    if err != nil {
        log.Fatal(err)
    }

    _, err = masterGroup.Register(ctx, "db1", mgorm.DBConfig{
        Name:            "主库-数据库1",
        DSN:             "user:password@tcp(master.db.example.com:3306)/db1?charset=utf8mb4&parseTime=True&loc=Local",
        Dialector:       mysql.Open("user:password@tcp(master.db.example.com:3306)/db1?charset=utf8mb4&parseTime=True&loc=Local"),
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 获取从库组并注册连接
    slaveGroup, err := manager.Group("slave")
    if err != nil {
        log.Fatal(err)
    }

    _, err = slaveGroup.Register(ctx, "db1", mgorm.DBConfig{
        Name:            "从库-数据库1",
        DSN:             "user:password@tcp(slave.db.example.com:3306)/db1?charset=utf8mb4&parseTime=True&loc=Local",
        Dialector:       mysql.Open("user:password@tcp(slave.db.example.com:3306)/db1?charset=utf8mb4&parseTime=True&loc=Local"),
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 写操作使用主库
    masterDB, _ := masterGroup.Get(ctx, "db1")
    masterDB.Exec("INSERT INTO users (name) VALUES (?)", "张三")

    // 读操作使用从库
    slaveDB, _ := slaveGroup.Get(ctx, "db1")
    var users []map[string]interface{}
    slaveDB.Raw("SELECT * FROM users").Scan(&users)

    // 程序退出时关闭所有连接
    defer manager.Close(ctx)
}
```

## API 参考

### DBConfig 配置项

| 字段              | 类型             | 说明                                  |
| ----------------- | ---------------- | ------------------------------------- |
| `Name`            | `string`         | 数据库描述名称（可选，用于日志记录）  |
| `DSN`             | `string`         | 数据源名称（连接字符串，可选）        |
| `DriverType`      | `string`         | 驱动类型（如 mysql, postgres 等）    |
| `Host`            | `string`         | 数据库主机地址                        |
| `Port`            | `int`            | 数据库端口                            |
| `User`            | `string`         | 数据库用户名                          |
| `Password`        | `string`         | 数据库密码                            |
| `DBName`          | `string`         | 数据库名称                            |
| `Charset`         | `string`         | 字符集（默认 utf8mb4）                |
| `Dialector`       | `gorm.Dialector` | GORM 方言驱动（**必需**，或使用自动生成） |
| `MaxIdleConns`    | `int`            | 最大空闲连接数                        |
| `MaxOpenConns`    | `int`            | 最大打开连接数                        |
| `ConnMaxLifetime` | `time.Duration`  | 连接最大存活时间                      |

### 单组管理 API

```go
// 创建单组数据库连接管理器
group := mgorm.New()

// 注册数据库连接
isNew, err := group.Register(ctx, "name", config)

// 获取数据库连接
db, err := group.Get(ctx, "name")

// 获取数据库连接（不存在时 panic）
db := group.MustGet(ctx, "name")

// 注销数据库连接
err := group.Unregister(ctx, "name")

// 获取所有已注册的连接名称列表
names := group.List()

// 关闭所有连接
errs := group.Close(ctx)
```

### 多组管理 API

```go
// 创建多组数据库连接管理器
manager := mgorm.NewManager()

// 添加资源组
existed := manager.AddGroup("groupName")

// 获取资源组
group, err := manager.Group("groupName")

// 获取资源组（不存在时 panic）
group := manager.MustGroup("groupName")

// 获取所有组名
names := manager.ListGroupNames()

// 关闭所有组的所有连接
errs := manager.Close(ctx)
```

## 完整示例：CRUD 操作

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/qq1060656096/mgorm"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    ID        uint           `gorm:"primarykey"`
    Name      string         `gorm:"size:100"`
    Email     string         `gorm:"size:255;uniqueIndex"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

func main() {
    ctx := context.Background()
    group := mgorm.New()

    // 注册数据库
    dsn := "user:password@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
    _, err := group.Register(ctx, "main", mgorm.DBConfig{
        Name:            "主数据库",
        DSN:             dsn,
        Dialector:       mysql.Open(dsn),
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer group.Close(ctx)

    // 获取数据库连接
    db, err := group.Get(ctx, "main")
    if err != nil {
        log.Fatal(err)
    }

    // 自动迁移
    db.AutoMigrate(&User{})

    // 创建用户
    user := User{Name: "张三", Email: "zhangsan@example.com"}
    result := db.Create(&user)
    if result.Error != nil {
        log.Fatal(result.Error)
    }
    fmt.Printf("创建用户成功，ID: %d\n", user.ID)

    // 查询用户
    var foundUser User
    db.First(&foundUser, user.ID)
    fmt.Printf("查询用户: %+v\n", foundUser)

    // 更新用户
    db.Model(&foundUser).Update("Name", "李四")
    fmt.Printf("更新后用户: %+v\n", foundUser)

    // 删除用户
    db.Delete(&foundUser)
    fmt.Println("用户已删除")
}
```

## 自动生成 DSN

mgorm 支持根据配置字段自动生成 DSN，无需手动编写连接字符串。

### 使用自动生成 DSN

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/qq1060656096/mgorm"
    "gorm.io/driver/mysql"
)

func main() {
    ctx := context.Background()
    group := mgorm.New()

    // 使用自动生成 DSN 配置 MySQL
    config := mgorm.DBConfig{
        Name:            "主数据库",
        DriverType:      "mysql",
        Host:            "127.0.0.1",
        Port:            3306,
        User:            "user",
        Password:        "password",
        DBName:          "testdb",
        Charset:         "utf8mb4", // 可选，默认 utf8mb4
        MaxIdleConns:    10,
        MaxOpenConns:    100,
        ConnMaxLifetime: time.Hour,
    }

    // 自动生成 DSN 并创建连接
    _, err := group.Register(ctx, "main", config)
    if err != nil {
        log.Fatal(err)
    }

    // 获取数据库连接
    db, err := group.Get(ctx, "main")
    if err != nil {
        log.Fatal(err)
    }

    // 使用数据库连接
    var result []map[string]interface{}
    db.Raw("SELECT 1").Scan(&result)

    defer group.Close(ctx)
}
```

### 支持的数据库类型

| 数据库类型 | `DriverType` 值 | 生成的 DSN 格式示例 |
| ---------- | --------------- | ------------------- |
| MySQL      | `mysql`         | `user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local` |
| PostgreSQL | `postgres`      | `host=host port=port user=user password=password dbname=dbname sslmode=disable` |
| SQLite     | `sqlite`        | `dbname`（直接使用文件路径） |
| SQL Server | `sqlserver`     | `sqlserver://user:password@host:port?database=dbname` |

### 优先级说明

1. **优先使用 `Dialector`**：如果设置了 `Dialector` 字段，将忽略其他 DSN 相关配置
2. **其次使用 `DSN`**：如果设置了 `DSN` 字段，将直接使用该值
3. **最后自动生成**：如果以上两者都未设置，将根据 `DriverType` 等字段自动生成 DSN

## MySQL DSN 格式

```
用户名:密码@tcp(主机:端口)/数据库名?参数
```

**常用参数：**

| 参数        | 说明                          | 推荐值    |
| ----------- | ----------------------------- | --------- |
| `charset`   | 字符集                        | `utf8mb4` |
| `parseTime` | 是否解析 `time.Time` 类型     | `True`    |
| `loc`       | 时区                          | `Local`   |
| `timeout`   | 连接超时                      | `10s`     |
| `readTimeout` | 读取超时                    | `30s`     |
| `writeTimeout` | 写入超时                   | `30s`     |

**完整示例：**

```
user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s
```

## 连接池配置建议

| 参数              | 说明           | 建议值       |
| ----------------- | -------------- | ------------ |
| `MaxIdleConns`    | 最大空闲连接数 | 10-25        |
| `MaxOpenConns`    | 最大打开连接数 | 100-200      |
| `ConnMaxLifetime` | 连接最大存活   | 1小时以内    |

> **注意**: `MaxIdleConns` 应小于等于 `MaxOpenConns`

## 支持的数据库

mgorm 基于 GORM，支持所有 GORM 支持的数据库：

| 数据库     | 驱动包                      |
| ---------- | --------------------------- |
| MySQL      | `gorm.io/driver/mysql`      |
| PostgreSQL | `gorm.io/driver/postgres`   |
| SQLite     | `gorm.io/driver/sqlite`     |
| SQL Server | `gorm.io/driver/sqlserver`  |
| ClickHouse | `gorm.io/driver/clickhouse` |

## 错误处理

```go
// 检查是否为缺少 DSN 错误
if mgorm.IsErrNoDSN(err) {
    log.Println("需要提供 DSN 配置")
}

// 检查是否为缺少 Dialector 错误
if mgorm.IsErrNoDialector(err) {
    log.Println("需要提供 Dialector 配置")
}
```

## 实现原理

mgorm 基于 [bizutil/registry](https://github.com/qq1060656096/bizutil/tree/main/registry) 包实现，该包提供了通用的资源注册与管理功能。

### 核心架构

```
┌─────────────────────────────────────────────────────────────┐
│                         mgorm                               │
├─────────────────────────────────────────────────────────────┤
│  New() -> registry.Group[DBConfig, *gorm.DB]                │
│  NewManager() -> registry.Manager[DBConfig, *gorm.DB]       │
├─────────────────────────────────────────────────────────────┤
│                   bizutil/registry                          │
│  ┌─────────────┐    ┌──────────────────────────────────┐    │
│  │   Group     │    │            Manager               │    │
│  │  (单组管理)  │    │  (多组管理，包含多个 Group)       │    │
│  └─────────────┘    └──────────────────────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                        GORM                                 │
│                    (数据库 ORM)                              │
└─────────────────────────────────────────────────────────────┘
```

### 关键函数

mgorm 通过实现 `opener` 和 `closer` 函数，将数据库连接的生命周期管理委托给 registry：

```go
// opener - 创建数据库连接
func opener(ctx context.Context, cfg DBConfig) (*gorm.DB, error) {
    // 1. 验证配置
    // 2. 使用 Dialector 打开连接
    // 3. 设置连接池参数
    // 4. Ping 验证连接可用
    return db, nil
}

// closer - 关闭数据库连接
func closer(ctx context.Context, db *gorm.DB) error {
    // 安全关闭底层 SQL 连接
    return sqlDB.Close()
}

// 创建单组管理器
func New() registry.Group[DBConfig, *gorm.DB] {
    return registry.NewGroup[DBConfig, *gorm.DB](opener, closer)
}

// 创建多组管理器
func NewManager() registry.Manager[DBConfig, *gorm.DB] {
    return registry.New[DBConfig, *gorm.DB](opener, closer)
}
```

### registry 包特性

- **惰性初始化**: 资源在首次 `Get()` 时才创建，而非注册时
- **线程安全**: 内部使用互斥锁保证并发安全
- **统一生命周期**: 通过 `Close()` 统一关闭所有资源
- **泛型支持**: 使用 Go 泛型，支持任意配置类型和资源类型

## 许可证

[Apache License](LICENSE)
