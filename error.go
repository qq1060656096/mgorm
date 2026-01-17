package mgorm

import (
	"errors"
)

// 定义包中的标准错误
var (
	// errNoDSN 表示未提供 DSN 配置
	errNoDSN = errors.New("mgorm: DSN is required when Dialector is not provided")
	// errNoDialector 表示需要提供 Dialector 或导入相应的驱动包
	errNoDialector = errors.New("mgorm: please provide a Dialector in DBConfig, or import the appropriate driver package")
)

// IsErrNoDSN 检查错误是否为缺少 DSN 配置错误
func IsErrNoDSN(err error) bool {
	return errors.Is(err, errNoDSN)
}

// IsErrNoDialector 检查错误是否为缺少 Dialector 错误
func IsErrNoDialector(err error) bool {
	return errors.Is(err, errNoDialector)
}
