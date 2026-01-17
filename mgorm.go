// Package mgorm 提供基于 GORM 的数据库连接管理功能。
// 该包封装了数据库连接的创建、配置和生命周期管理，
// 支持连接池配置、多数据库实例管理等功能。
package mgorm

import (
	"context"

	"github.com/qq1060656096/bizutil/registry"
	"gorm.io/gorm"
)

// opener 根据配置创建并初始化数据库连接。
// 该函数会执行以下操作：
//   - 验证数据库配置的有效性
//   - 使用配置的 Dialector 打开数据库连接
//   - 设置连接池参数（最大空闲连接数、最大打开连接数、连接最大存活时间）
//   - 通过 Ping 验证数据库连接是否可用
//
// 参数：
//   - ctx: 上下文，用于控制连接超时
//   - cfg: 数据库配置信息
//
// 返回：
//   - *gorm.DB: 成功时返回 GORM 数据库实例
//   - error: 配置验证失败、连接失败或 Ping 失败时返回错误
func opener(ctx context.Context, cfg DBConfig) (*gorm.DB, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	db, err := gorm.Open(cfg.Dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	return db, nil
}

// closer 关闭数据库连接。
// 该函数会安全地关闭 GORM 数据库实例底层的 SQL 连接。
// 如果传入的 db 为 nil，则直接返回 nil 不执行任何操作。
//
// 参数：
//   - ctx: 上下文（当前未使用，预留用于未来扩展）
//   - db: 需要关闭的 GORM 数据库实例
//
// 返回：
//   - error: 获取底层连接失败或关闭连接失败时返回错误
func closer(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Group 是单一组管理（key => redis client）
type Group = registry.Group[DBConfig, *gorm.DB]

// Manager 是多组管理
type Manager = registry.Manager[DBConfig, *gorm.DB]

// NewManager 创建一个新的数据库连接管理器。
// 返回的 Manager 实例用于管理单个数据库连接的生命周期，
// 包括连接的创建、获取和关闭。
//
// 返回：
//   - registry.Manager[DBConfig, *gorm.DB]: 数据库连接管理器实例
func NewManager() Manager {
	return registry.New[DBConfig, *gorm.DB](
		opener,
		closer,
	)
}

// New 创建一个新的数据库连接分组管理器。
// 返回的 Group 实例用于管理多个命名的数据库连接，
// 适用于需要同时管理多个数据库实例的场景（如主从分离、多租户等）。
//
// 返回：
//   - registry.Group[DBConfig, *gorm.DB]: 数据库连接分组管理器实例
func New() Group {
	return registry.NewGroup[DBConfig, *gorm.DB](
		opener,
		closer,
	)
}
