package mgorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// RegisterToDB 使用当前 Group 中已有名称 fromName 的配置，
// 将其注册为新的名称 toName，并写入指定数据库 toDBName。
// 返回值 isNew 表示 toName 是否为新注册。
func RegisterToDB(ctx context.Context, group Group, fromName, toName, toDBName string) (isNew bool, err error) {
	cfg, err := group.Config(ctx, fromName)
	if err != nil {
		return false, err
	}

	cfg.Name = toName
	cfg.DBName = toDBName
	cfg.DSN = ""
	cfg.DSN = cfg.AutoDsn()
	dialector, err := CreateDialector(cfg.DriverType, cfg.DSN)
	if err != nil {
		return false, err
	}
	cfg.Dialector = dialector

	return group.Register(ctx, toName, cfg)
}

// MustRegisterToDB 使用当前 Group 中已有名称 fromName 的配置，
// 将其注册为新的名称 toName，并写入指定数据库 toDBName。
// 返回值 isNew 表示 toName 是否为新注册。
func MustRegisterToDB(ctx context.Context, group Group, fromName, toName, toDBName string) (isNew bool) {
	isNew, err := RegisterToDB(ctx, group, fromName, toName, toDBName)
	if err != nil {
		panic(err)
	}
	return isNew
}

// BatchMustRegisterToDB 批量将一个来源 DB(fromName)
// 注册到同一个 Group 下的多个目标 DB。
//
// 该方法是 MustRegisterToDB 的批量封装版本，
// 内部会遍历 toNameDBMap 并逐个调用 MustRegisterToDB。
// 任一注册失败都会直接 panic（由 MustRegisterToDB 保证）。
//
// 参数说明：
//   - ctx:        上下文，用于生命周期控制
//   - group:      DB 分组（如 public / business）
//   - fromName:   来源 DB 名称（通常为 default）
//   - toNameDBMap:
//     key   -> 目标逻辑名 toName（如 order / goods）
//     value -> 实际物理 DB 名称 toDBName（如 data_1 / data_2）
func BatchMustRegisterToDB(ctx context.Context, group Group, fromName string, toNameDBMap map[string]string) {
	for toName, toDBName := range toNameDBMap {
		MustRegisterToDB(ctx, group, fromName, toName, toDBName)
	}
}

// ErrUnknownDriverType 当指定了不支持的数据库驱动类型时返回此错误。
var ErrUnknownDriverType = errors.New("mgorm: unknown driver type")

func CreateDialector(driverType, dsn string) (gorm.Dialector, error) {
	switch driverType {
	case "mysql":
		return mysql.Open(dsn), nil
	case "postgres":
		return postgres.Open(dsn), nil
	case "sqlite":
		return sqlite.Open(dsn), nil
	case "sqlserver":
		return sqlserver.Open(dsn), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownDriverType, driverType)
	}
}
