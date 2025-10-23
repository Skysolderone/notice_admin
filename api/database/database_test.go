package database

import (
	"testing"
	"time"

	"notice/api/config"

	"gorm.io/gorm"
)

// TestModel 测试用的数据模型
type TestModel struct {
	gorm.Model
	Name        string `gorm:"size:100;not null"`
	Description string `gorm:"size:500"`
	Value       int
}

// TestInitDB 测试数据库初始化
func TestInitDB(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.DatabaseConfig
		wantErr bool
	}{
		{
			name: "有效配置",
			cfg: config.DatabaseConfig{
				Host:            "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
				Port:            5432,
				User:            "wws",
				Password:        "Wws5201314",
				DBName:          "notice",
				SSLMode:         "disable",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: 3600,
			},
			wantErr: false,
		},
		{
			name: "使用默认端口",
			cfg: config.DatabaseConfig{
				Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
				User:     "wws",
				Password: "Wws5201314",
				DBName:   "notice",
			},
			wantErr: false,
		},
		{
			name: "错误的主机地址",
			cfg: config.DatabaseConfig{
				Host:     "invalid_host_12345",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				DBName:   "test_db",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
		{
			name: "错误的端口",
			cfg: config.DatabaseConfig{
				Host:     "localhost",
				Port:     99999,
				User:     "postgres",
				Password: "postgres",
				DBName:   "test_db",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InitDB(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// 验证数据库实例已创建
				if db == nil {
					t.Error("InitDB() 成功但 db 实例为 nil")
					return
				}

				// 验证连接池参数
				sqlDB, err := db.DB()
				if err != nil {
					t.Errorf("获取 sql.DB 失败: %v", err)
					return
				}

				// 测试 Ping
				if err := sqlDB.Ping(); err != nil {
					t.Errorf("数据库 Ping 失败: %v", err)
				}

				// 验证连接池设置
				expectedMaxOpen := tt.cfg.MaxOpenConns
				if expectedMaxOpen == 0 {
					expectedMaxOpen = 10
				}
				stats := sqlDB.Stats()
				if stats.MaxOpenConnections != expectedMaxOpen {
					t.Errorf("MaxOpenConns = %d, want %d", stats.MaxOpenConnections, expectedMaxOpen)
				}

				// 清理
				CloseDB()
			}
		})
	}
}

// TestGetDB 测试获取数据库实例
func TestGetDB(t *testing.T) {
	// 初始化数据库
	cfg := config.DatabaseConfig{
		Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:     5432,
		User:     "wws",
		Password: "Wws5201314",
		DBName:   "notice",
		SSLMode:  "disable",
	}

	err := InitDB(cfg)
	if err != nil {
		t.Skipf("跳过测试 - 无法连接数据库: %v", err)
		return
	}
	defer CloseDB()

	// 测试获取实例
	instance := GetDB()
	if instance == nil {
		t.Error("GetDB() 返回 nil")
		return
	}

	// 验证实例可用
	sqlDB, err := instance.DB()
	if err != nil {
		t.Errorf("获取 sql.DB 失败: %v", err)
		return
	}

	if err := sqlDB.Ping(); err != nil {
		t.Errorf("数据库 Ping 失败: %v", err)
	}
}

// TestCloseDB 测试关闭数据库连接
func TestCloseDB(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:     5432,
		User:     "wws",
		Password: "Wws5201314",
		DBName:   "notice",
		SSLMode:  "disable",
	}

	err := InitDB(cfg)
	if err != nil {
		t.Skipf("跳过测试 - 无法连接数据库: %v", err)
		return
	}

	// 测试关闭
	err = CloseDB()
	if err != nil {
		t.Errorf("CloseDB() error = %v", err)
	}

	// 测试重复关闭（应该返回 nil）
	err = CloseDB()
	if err != nil {
		t.Errorf("重复调用 CloseDB() error = %v", err)
	}
}

// TestAutoMigrate 测试自动迁移
func TestAutoMigrate(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:     5432,
		User:     "wws",
		Password: "Wws5201314",
		DBName:   "notice",
		SSLMode:  "disable",
	}

	err := InitDB(cfg)
	if err != nil {
		t.Skipf("跳过测试 - 无法连接数据库: %v", err)
		return
	}
	defer CloseDB()

	// 测试迁移
	err = AutoMigrate(&TestModel{})
	if err != nil {
		t.Errorf("AutoMigrate() error = %v", err)
		return
	}

	// 验证表已创建
	if !db.Migrator().HasTable(&TestModel{}) {
		t.Error("AutoMigrate() 后表未创建")
	}

	// 清理测试表
	db.Migrator().DropTable(&TestModel{})
}

// TestDatabaseCRUD 测试基本的 CRUD 操作
func TestDatabaseCRUD(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:     5432,
		User:     "wws",
		Password: "Wws5201314",
		DBName:   "notice",
		SSLMode:  "disable",
	}

	err := InitDB(cfg)
	if err != nil {
		t.Skipf("跳过测试 - 无法连接数据库: %v", err)
		return
	}
	defer CloseDB()

	// 迁移测试表
	if err := AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	defer db.Migrator().DropTable(&TestModel{})

	t.Run("创建记录", func(t *testing.T) {
		model := TestModel{
			Name:        "测试记录",
			Description: "这是一条测试记录",
			Value:       100,
		}

		result := db.Create(&model)
		if result.Error != nil {
			t.Errorf("创建记录失败: %v", result.Error)
			return
		}

		if model.ID == 0 {
			t.Error("创建记录后 ID 为 0")
		}
	})

	t.Run("查询记录", func(t *testing.T) {
		var models []TestModel
		result := db.Find(&models)
		if result.Error != nil {
			t.Errorf("查询记录失败: %v", result.Error)
			return
		}

		if len(models) == 0 {
			t.Error("查询结果为空")
		}
	})

	t.Run("更新记录", func(t *testing.T) {
		var model TestModel
		db.First(&model)

		model.Value = 200
		result := db.Save(&model)
		if result.Error != nil {
			t.Errorf("更新记录失败: %v", result.Error)
			return
		}

		// 验证更新
		var updated TestModel
		db.First(&updated, model.ID)
		if updated.Value != 200 {
			t.Errorf("更新后的值 = %d, want 200", updated.Value)
		}
	})

	t.Run("删除记录", func(t *testing.T) {
		var model TestModel
		db.First(&model)

		result := db.Delete(&model)
		if result.Error != nil {
			t.Errorf("删除记录失败: %v", result.Error)
			return
		}

		// 验证删除
		var count int64
		db.Model(&TestModel{}).Where("id = ?", model.ID).Count(&count)
		if count != 0 {
			t.Error("删除后记录仍然存在")
		}
	})
}

// TestConnectionPool 测试连接池功能
func TestConnectionPool(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:            "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:            5432,
		User:            "wws",
		Password:        "",
		DBName:          "notice",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 60,
	}

	err := InitDB(cfg)
	if err != nil {
		t.Skipf("跳过测试 - 无法连接数据库: %v", err)
		return
	}
	defer CloseDB()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("获取 sql.DB 失败: %v", err)
	}

	// 验证连接池配置
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != cfg.MaxOpenConns {
		t.Errorf("MaxOpenConnections = %d, want %d", stats.MaxOpenConnections, cfg.MaxOpenConns)
	}

	// 执行一些查询来使用连接
	for i := 0; i < 10; i++ {
		var result int
		db.Raw("SELECT 1").Scan(&result)
	}

	// 等待一段时间让连接回收
	time.Sleep(100 * time.Millisecond)

	// 检查连接统计
	stats = sqlDB.Stats()
	t.Logf("连接统计 - 打开: %d, 使用中: %d, 空闲: %d",
		stats.OpenConnections, stats.InUse, stats.Idle)

	if stats.OpenConnections > cfg.MaxOpenConns {
		t.Errorf("打开的连接数 %d 超过最大值 %d", stats.OpenConnections, cfg.MaxOpenConns)
	}
}

// TestTransactions 测试事务功能
func TestTransactions(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:     5432,
		User:     "wws",
		Password: "Wws5201314",
		DBName:   "notice",
		SSLMode:  "disable",
	}

	err := InitDB(cfg)
	if err != nil {
		t.Skipf("跳过测试 - 无法连接数据库: %v", err)
		return
	}
	defer CloseDB()

	if err := AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	defer db.Migrator().DropTable(&TestModel{})

	t.Run("事务提交", func(t *testing.T) {
		err := db.Transaction(func(tx *gorm.DB) error {
			model := TestModel{Name: "事务测试", Value: 100}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			t.Errorf("事务提交失败: %v", err)
			return
		}

		// 验证记录已创建
		var count int64
		db.Model(&TestModel{}).Where("name = ?", "事务测试").Count(&count)
		if count != 1 {
			t.Errorf("事务提交后记录数 = %d, want 1", count)
		}
	})

	t.Run("事务回滚", func(t *testing.T) {
		initialCount := int64(0)
		db.Model(&TestModel{}).Count(&initialCount)

		err := db.Transaction(func(tx *gorm.DB) error {
			model := TestModel{Name: "回滚测试", Value: 200}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
			// 模拟错误，触发回滚
			return gorm.ErrInvalidTransaction
		})

		if err == nil {
			t.Error("期望事务失败但成功了")
			return
		}

		// 验证记录未创建
		var finalCount int64
		db.Model(&TestModel{}).Count(&finalCount)
		if finalCount != initialCount {
			t.Errorf("事务回滚后记录数 = %d, want %d", finalCount, initialCount)
		}
	})
}

// BenchmarkDatabaseQuery 性能测试 - 查询
func BenchmarkDatabaseQuery(b *testing.B) {
	cfg := config.DatabaseConfig{
		Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:     5432,
		User:     "wws",
		Password: "Wws5201314",
		DBName:   "notice",
		SSLMode:  "disable",
	}

	err := InitDB(cfg)
	if err != nil {
		b.Skipf("跳过测试 - 无法连接数据库: %v", err)
		return
	}
	defer CloseDB()

	if err := AutoMigrate(&TestModel{}); err != nil {
		b.Fatalf("迁移失败: %v", err)
	}
	defer db.Migrator().DropTable(&TestModel{})

	// 插入测试数据
	for i := 0; i < 100; i++ {
		db.Create(&TestModel{Name: "性能测试", Value: i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var models []TestModel
		db.Limit(10).Find(&models)
	}
}

// BenchmarkDatabaseInsert 性能测试 - 插入
func BenchmarkDatabaseInsert(b *testing.B) {
	cfg := config.DatabaseConfig{
		Host:     "pgm-bp140jpn9wct9u0t.pg.rds.aliyuncs.com",
		Port:     5432,
		User:     "wws",
		Password: "Wws5201314",
		DBName:   "notice",
		SSLMode:  "disable",
	}

	err := InitDB(cfg)
	if err != nil {
		b.Skipf("跳过测试 - 无法连接数据库: %v", err)
		return
	}
	defer CloseDB()

	if err := AutoMigrate(&TestModel{}); err != nil {
		b.Fatalf("迁移失败: %v", err)
	}
	defer db.Migrator().DropTable(&TestModel{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model := TestModel{Name: "性能测试", Value: i}
		db.Create(&model)
	}
}
