package mgorm

import (
	"context"
	"testing"
)

// ==================== RegisterToDB 测试 ====================

// TestRegisterToDB 测试 RegisterToDB 函数的正常功能
func TestRegisterToDB(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 准备源配置
	sourceConfig := DBConfig{
		Name:       "源数据库",
		DriverType: "sqlite",
		Host:       "",
		Port:       0,
		User:       "",
		Password:   "",
		DBName:     ":memory:",
	}

	// 注册源数据库
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		t.Fatalf("注册源数据库失败: %v", err)
	}

	// 使用 RegisterToDB 注册目标数据库
	isNew, err := RegisterToDB(ctx, group, "source", "target", "target_db")
	if err != nil {
		t.Fatalf("RegisterToDB 失败: %v", err)
	}
	if !isNew {
		t.Error("应该是新注册的数据库")
	}

	// 验证目标数据库配置
	targetConfig, err := group.Config(ctx, "target")
	if err != nil {
		t.Fatalf("获取目标配置失败: %v", err)
	}
	if targetConfig.Name != "target" {
		t.Errorf("目标配置 Name = %q, 期望 %q", targetConfig.Name, "target")
	}
	if targetConfig.DBName != "target_db" {
		t.Errorf("目标配置 DBName = %q, 期望 %q", targetConfig.DBName, "target_db")
	}

	// 验证目标数据库连接可用
	targetDB, err := group.Get(ctx, "target")
	if err != nil {
		t.Fatalf("获取目标数据库连接失败: %v", err)
	}
	if targetDB == nil {
		t.Fatal("目标数据库连接不应为 nil")
	}

	// 清理
	group.Close(ctx)
}

// TestRegisterToDB_ExistingTarget 测试注册已存在的目标数据库名称
func TestRegisterToDB_ExistingTarget(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		t.Fatalf("注册源数据库失败: %v", err)
	}

	// 注册目标数据库
	targetConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err = group.Register(ctx, "target", targetConfig)
	if err != nil {
		t.Fatalf("注册目标数据库失败: %v", err)
	}

	// 再次使用 RegisterToDB 注册相同的目标名称
	isNew, err := RegisterToDB(ctx, group, "source", "target", "new_target_db")
	if err != nil {
		t.Fatalf("RegisterToDB 失败: %v", err)
	}
	if isNew {
		t.Error("目标数据库已存在，不应返回 isNew = true")
	}

	// 清理
	group.Close(ctx)
}

// TestRegisterToDB_NonExistentSource 测试使用不存在的源数据库名称
func TestRegisterToDB_NonExistentSource(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 尝试使用不存在的源数据库
	isNew, err := RegisterToDB(ctx, group, "nonexistent", "target", "target_db")
	if err == nil {
		t.Error("使用不存在的源数据库应返回错误")
	}
	if isNew {
		t.Error("失败时 isNew 应为 false")
	}

	group.Close(ctx)
}

// TestRegisterToDB_EmptyNames 测试使用空名称参数
func TestRegisterToDB_EmptyNames(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		t.Fatalf("注册源数据库失败: %v", err)
	}

	// 测试空目标数据库名（DBName 可以为空）
	isNew, err := RegisterToDB(ctx, group, "source", "target", "")
	if err != nil {
		t.Fatalf("RegisterToDB 失败: %v", err)
	}
	if !isNew {
		t.Error("应该是新注册的数据库")
	}

	// 验证配置
	targetConfig, err := group.Config(ctx, "target")
	if err != nil {
		t.Fatalf("获取目标配置失败: %v", err)
	}
	if targetConfig.DBName != "" {
		t.Errorf("目标配置 DBName = %q, 期望空字符串", targetConfig.DBName)
	}

	group.Close(ctx)
}

// ==================== MustRegisterToDB 测试 ====================

// TestMustRegisterToDB 测试 MustRegisterToDB 函数的正常功能
func TestMustRegisterToDB(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		t.Fatalf("注册源数据库失败: %v", err)
	}

	// 使用 MustRegisterToDB 注册目标数据库
	isNew := MustRegisterToDB(ctx, group, "source", "target", "target_db")
	if !isNew {
		t.Error("应该是新注册的数据库")
	}

	// 验证目标数据库存在
	_, err = group.Get(ctx, "target")
	if err != nil {
		t.Fatalf("获取目标数据库失败: %v", err)
	}

	// 清理
	group.Close(ctx)
}

// TestMustRegisterToDB_Panic 测试 MustRegisterToDB 在错误时是否 panic
func TestMustRegisterToDB_Panic(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 测试使用不存在的源数据库应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustRegisterToDB 使用不存在的源数据库应该 panic")
		}
	}()

	MustRegisterToDB(ctx, group, "nonexistent", "target", "target_db")
}

// TestMustRegisterToDB_ExistingTarget 测试 MustRegisterToDB 处理已存在目标
func TestMustRegisterToDB_ExistingTarget(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		t.Fatalf("注册源数据库失败: %v", err)
	}

	// 注册目标数据库
	targetConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err = group.Register(ctx, "target", targetConfig)
	if err != nil {
		t.Fatalf("注册目标数据库失败: %v", err)
	}

	// 使用 MustRegisterToDB 注册已存在的目标
	isNew := MustRegisterToDB(ctx, group, "source", "target", "new_target_db")
	if isNew {
		t.Error("目标数据库已存在，不应返回 isNew = true")
	}

	// 清理
	group.Close(ctx)
}

// ==================== BatchMustRegisterToDB 测试 ====================

// TestBatchMustRegisterToDB 测试 BatchMustRegisterToDB 函数的正常功能
func TestBatchMustRegisterToDB(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		t.Fatalf("注册源数据库失败: %v", err)
	}

	// 批量注册目标数据库
	toNameDBMap := map[string]string{
		"order":   "order_db",
		"goods":   "goods_db",
		"user":    "user_db",
		"payment": "payment_db",
	}

	BatchMustRegisterToDB(ctx, group, "source", toNameDBMap)

	// 验证所有目标数据库都已注册
	for toName, toDBName := range toNameDBMap {
		// 验证配置
		config, err := group.Config(ctx, toName)
		if err != nil {
			t.Fatalf("获取 %s 配置失败: %v", toName, err)
		}
		if config.Name != toName {
			t.Errorf("%s 配置 Name = %q, 期望 %q", toName, config.Name, toName)
		}
		if config.DBName != toDBName {
			t.Errorf("%s 配置 DBName = %q, 期望 %q", toName, config.DBName, toDBName)
		}

		// 验证连接可用
		db, err := group.Get(ctx, toName)
		if err != nil {
			t.Fatalf("获取 %s 数据库连接失败: %v", toName, err)
		}
		if db == nil {
			t.Errorf("%s 数据库连接不应为 nil", toName)
		}
	}

	// 验证资源列表
	list := group.List()
	expectedCount := 1 + len(toNameDBMap) // source + targets
	if len(list) != expectedCount {
		t.Errorf("期望 %d 个资源，实际 %d 个", expectedCount, len(list))
	}

	// 清理
	group.Close(ctx)
}

// TestBatchMustRegisterToDB_EmptyMap 测试批量注册空映射
func TestBatchMustRegisterToDB_EmptyMap(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		t.Fatalf("注册源数据库失败: %v", err)
	}

	// 批量注册空映射（应该不执行任何操作）
	BatchMustRegisterToDB(ctx, group, "source", map[string]string{})

	// 验证只有源数据库存在
	list := group.List()
	if len(list) != 1 {
		t.Errorf("期望 1 个资源，实际 %d 个", len(list))
	}

	// 清理
	group.Close(ctx)
}

// TestBatchMustRegisterToDB_Panic 测试 BatchMustRegisterToDB 在错误时是否 panic
func TestBatchMustRegisterToDB_Panic(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 测试使用不存在的源数据库应该 panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("BatchMustRegisterToDB 使用不存在的源数据库应该 panic")
		}
	}()

	toNameDBMap := map[string]string{
		"target": "target_db",
	}

	BatchMustRegisterToDB(ctx, group, "nonexistent", toNameDBMap)
}

// TestBatchMustRegisterToDB_SingleTarget 测试批量注册单个目标
func TestBatchMustRegisterToDB_SingleTarget(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		t.Fatalf("注册源数据库失败: %v", err)
	}

	// 批量注册单个目标
	toNameDBMap := map[string]string{
		"single": "single_db",
	}

	BatchMustRegisterToDB(ctx, group, "source", toNameDBMap)

	// 验证目标数据库已注册
	config, err := group.Config(ctx, "single")
	if err != nil {
		t.Fatalf("获取 single 配置失败: %v", err)
	}
	if config.Name != "single" {
		t.Errorf("single 配置 Name = %q, 期望 %q", config.Name, "single")
	}
	if config.DBName != "single_db" {
		t.Errorf("single 配置 DBName = %q, 期望 %q", config.DBName, "single_db")
	}

	// 清理
	group.Close(ctx)
}

// ==================== 集成测试 ====================

// TestFunc_Integration 测试所有函数的集成使用
func TestFunc_Integration(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册多个源数据库
	sourceConfigs := map[string]DBConfig{
		"master": {
			Name:       "主数据库",
			DriverType: "sqlite",
			DBName:     ":memory:",
		},
		"slave": {
			Name:       "从数据库",
			DriverType: "sqlite",
			DBName:     ":memory:",
		},
	}

	for name, config := range sourceConfigs {
		_, err := group.Register(ctx, name, config)
		if err != nil {
			t.Fatalf("注册源数据库 %s 失败: %v", name, err)
		}
	}

	// 使用 RegisterToDB 从 master 注册目标
	isNew, err := RegisterToDB(ctx, group, "master", "order_db", "order_physical")
	if err != nil {
		t.Fatalf("RegisterToDB 失败: %v", err)
	}
	if !isNew {
		t.Error("应该是新注册的数据库")
	}

	// 使用 MustRegisterToDB 从 slave 注册目标
	isNew = MustRegisterToDB(ctx, group, "slave", "user_db", "user_physical")
	if !isNew {
		t.Error("应该是新注册的数据库")
	}

	// 使用 BatchMustRegisterToDB 从 master 批量注册
	batchMap := map[string]string{
		"goods_db":   "goods_physical",
		"payment_db": "payment_physical",
	}
	BatchMustRegisterToDB(ctx, group, "master", batchMap)

	// 验证所有数据库都已注册（跳过可能被修改的源数据库）
	expectedDBs := []string{"order_db", "user_db", "goods_db", "payment_db"}
	for _, dbName := range expectedDBs {
		db, err := group.Get(ctx, dbName)
		if err != nil {
			t.Fatalf("获取数据库 %s 失败: %v", dbName, err)
		}
		if db == nil {
			t.Errorf("数据库 %s 连接不应为 nil", dbName)
		}
	}

	// 验证资源总数（至少应该有目标数据库）
	list := group.List()
	if len(list) < len(expectedDBs) {
		t.Errorf("期望至少 %d 个资源，实际 %d 个", len(expectedDBs), len(list))
	}

	// 清理
	group.Close(ctx)
}

// TestFunc_ConfigInheritance 测试配置继承功能
func TestFunc_ConfigInheritance(t *testing.T) {
	ctx := context.Background()
	group := New()

	// 注册带有详细配置的源数据库
	sourceConfig := DBConfig{
		Name:            "源数据库",
		DriverType:      "sqlite",
		DBName:          ":memory:",
		MaxIdleConns:    10,
		MaxOpenConns:    20,
		ConnMaxLifetime: 0, // 测试零值处理
	}

	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		t.Fatalf("注册源数据库失败: %v", err)
	}

	// 使用 RegisterToDB 注册目标数据库
	isNew, err := RegisterToDB(ctx, group, "source", "target", "target_db")
	if err != nil {
		t.Fatalf("RegisterToDB 失败: %v", err)
	}
	if !isNew {
		t.Error("应该是新注册的数据库")
	}

	// 验证目标配置继承了源配置（除了 Name 和 DBName）
	targetConfig, err := group.Config(ctx, "target")
	if err != nil {
		t.Fatalf("获取目标配置失败: %v", err)
	}

	// 验证 Name 和 DBName 被正确覆盖
	if targetConfig.Name != "target" {
		t.Errorf("目标配置 Name = %q, 期望 %q", targetConfig.Name, "target")
	}
	if targetConfig.DBName != "target_db" {
		t.Errorf("目标配置 DBName = %q, 期望 %q", targetConfig.DBName, "target_db")
	}

	// 验证其他配置被继承
	if targetConfig.MaxIdleConns != sourceConfig.MaxIdleConns {
		t.Errorf("目标配置 MaxIdleConns = %d, 期望 %d", targetConfig.MaxIdleConns, sourceConfig.MaxIdleConns)
	}
	if targetConfig.MaxOpenConns != sourceConfig.MaxOpenConns {
		t.Errorf("目标配置 MaxOpenConns = %d, 期望 %d", targetConfig.MaxOpenConns, sourceConfig.MaxOpenConns)
	}
	if targetConfig.ConnMaxLifetime != sourceConfig.ConnMaxLifetime {
		t.Errorf("目标配置 ConnMaxLifetime = %v, 期望 %v", targetConfig.ConnMaxLifetime, sourceConfig.ConnMaxLifetime)
	}

	// 验证 Dialector 被重新创建（不是直接继承）
	if targetConfig.Dialector == nil {
		t.Error("目标配置 Dialector 不应为 nil")
	}

	// 清理
	group.Close(ctx)
}

// ==================== 性能测试 ====================

// BenchmarkRegisterToDB 性能测试 RegisterToDB 函数
func BenchmarkRegisterToDB(b *testing.B) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		b.Fatalf("注册源数据库失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 使用不同的目标名称避免冲突
		toName := "target_" + string(rune(i))
		toDBName := "target_db_" + string(rune(i))
		_, err := RegisterToDB(ctx, group, "source", toName, toDBName)
		if err != nil {
			b.Fatalf("RegisterToDB 失败: %v", err)
		}
	}

	// 清理
	group.Close(ctx)
}

// BenchmarkMustRegisterToDB 性能测试 MustRegisterToDB 函数
func BenchmarkMustRegisterToDB(b *testing.B) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		b.Fatalf("注册源数据库失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 使用不同的目标名称避免冲突
		toName := "target_" + string(rune(i))
		toDBName := "target_db_" + string(rune(i))
		MustRegisterToDB(ctx, group, "source", toName, toDBName)
	}

	// 清理
	group.Close(ctx)
}

// BenchmarkBatchMustRegisterToDB 性能测试 BatchMustRegisterToDB 函数
func BenchmarkBatchMustRegisterToDB(b *testing.B) {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	_, err := group.Register(ctx, "source", sourceConfig)
	if err != nil {
		b.Fatalf("注册源数据库失败: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 创建批量映射
		toNameDBMap := make(map[string]string)
		for j := 0; j < 10; j++ {
			toName := "target_" + string(rune(i*10+j))
			toDBName := "target_db_" + string(rune(i*10+j))
			toNameDBMap[toName] = toDBName
		}
		BatchMustRegisterToDB(ctx, group, "source", toNameDBMap)
	}

	// 清理
	group.Close(ctx)
}

// ==================== 示例 ====================

// ExampleRegisterToDB 展示 RegisterToDB 的使用方法
func ExampleRegisterToDB() {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		Name:       "主数据库",
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	group.Register(ctx, "master", sourceConfig)

	// 将主数据库配置注册为订单数据库
	isNew, err := RegisterToDB(ctx, group, "master", "order", "order_db")
	if err != nil {
		panic(err)
	}

	// 使用订单数据库连接
	orderDB, _ := group.Get(ctx, "order")
	_ = orderDB
	_ = isNew

	// 关闭连接
	group.Close(ctx)
}

// ExampleMustRegisterToDB 展示 MustRegisterToDB 的使用方法
func ExampleMustRegisterToDB() {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		Name:       "主数据库",
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	group.Register(ctx, "master", sourceConfig)

	// 将主数据库配置注册为用户数据库（失败时会 panic）
	isNew := MustRegisterToDB(ctx, group, "master", "user", "user_db")

	// 使用用户数据库连接
	userDB, _ := group.Get(ctx, "user")
	_ = userDB
	_ = isNew

	// 关闭连接
	group.Close(ctx)
}

// ExampleBatchMustRegisterToDB 展示 BatchMustRegisterToDB 的使用方法
func ExampleBatchMustRegisterToDB() {
	ctx := context.Background()
	group := New()

	// 注册源数据库
	sourceConfig := DBConfig{
		Name:       "主数据库",
		DriverType: "sqlite",
		DBName:     ":memory:",
	}
	group.Register(ctx, "master", sourceConfig)

	// 批量注册多个业务数据库
	toNameDBMap := map[string]string{
		"order":   "order_physical_db",
		"goods":   "goods_physical_db",
		"user":    "user_physical_db",
		"payment": "payment_physical_db",
	}

	BatchMustRegisterToDB(ctx, group, "master", toNameDBMap)

	// 使用各个业务数据库连接
	orderDB, _ := group.Get(ctx, "order")
	goodsDB, _ := group.Get(ctx, "goods")
	userDB, _ := group.Get(ctx, "user")
	paymentDB, _ := group.Get(ctx, "payment")

	_ = orderDB
	_ = goodsDB
	_ = userDB
	_ = paymentDB

	// 关闭连接
	group.Close(ctx)
}
