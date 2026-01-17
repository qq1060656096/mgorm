package mgorm

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// DBConfig 数据库配置
type DBConfig struct {
	Name            string         `yaml:"name" mapstructure:"name"`                           // 数据库描述名称（可选，用于日志记录等，不作为连接标识）
	DSN             string         `yaml:"dsn" mapstructure:"dsn"`                             // 数据源名称（连接字符串）
	DriverType      string         `yaml:"driver_type" mapstructure:"driver_type"`             // 驱动类型（如 mysql, postgres 等）
	MaxIdleConns    int            `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`       // 最大空闲连接数
	MaxOpenConns    int            `yaml:"max_open_conns" mapstructure:"max_open_conns"`       // 最大打开连接数
	ConnMaxLifetime time.Duration  `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"` // 连接最大生存时间
	Dialector       gorm.Dialector `yaml:"-" mapstructure:"-"`                                 // 自定义方言驱动（可选，如果设置则忽略 DriverType 和 DSN）
}

// Validate 验证数据库配置是否有效
func (c *DBConfig) Validate() error {
	// 如果没有提供 Dialector，则必须提供 DSN
	if c.Dialector == nil {
		if c.DSN == "" {
			return errNoDSN
		}
		return errNoDialector
	}
	return nil
}

// dbConnection 数据库连接信息
type dbConnection struct {
	db     *gorm.DB // 数据库连接实例
	config DBConfig // 数据库配置信息
}

// openDB 根据配置创建数据库连接
func openDB(cfg DBConfig) (*gorm.DB, error) {
	var dialector gorm.Dialector

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// 使用自定义 Dialector
	dialector = cfg.Dialector

	// 打开数据库连接
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 获取底层的 sql.DB 以设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 配置连接池
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
