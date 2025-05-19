package dao

import (
	"time"

	"gil_teacher/app/conf"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// PostgreSQLClient PostgreSQL客户端
type PostgreSQLClient struct {
	Read  *gorm.DB
	Write *gorm.DB
	log   *log.Helper
}

// NewPostgreSQLClient 创建PostgreSQL读写分离客户端
func NewPostgreSQLClient(c *conf.Data, logger log.Logger) (*PostgreSQLClient, func(), error) {
	_log := log.NewHelper(log.With(logger, "x_module", "dao/NewPostgreSQLClient"))

	var (
		readDB  *gorm.DB
		writeDB *gorm.DB
		err     error
	)

	// 配置GORM - 静默模式
	gormConfig := &gorm.Config{
		Logger:         gormlogger.Default.LogMode(gormlogger.Info),
		TranslateError: true,
	}

	// 初始化写库连接
	if c.PostgreSQLWrite != nil {
		writeDB, err = gorm.Open(postgres.Open(c.PostgreSQLWrite.Source), gormConfig)
		if err != nil {
			_log.Errorf("连接PostgreSQL写库失败: %v", err)
			return nil, nil, err
		}

		// 获取原生连接池并设置连接参数
		sqlDB, err := writeDB.DB()
		if err != nil {
			_log.Errorf("获取PostgreSQL写库原生连接失败: %v", err)
			return nil, nil, err
		}

		sqlDB.SetMaxIdleConns(c.PostgreSQLWrite.MaxIdleConns)
		sqlDB.SetMaxOpenConns(c.PostgreSQLWrite.MaxOpenConns)
		sqlDB.SetConnMaxIdleTime(time.Duration(c.PostgreSQLWrite.ConnMaxIdleTime) * time.Second)
		sqlDB.SetConnMaxLifetime(time.Duration(c.PostgreSQLWrite.ConnMaxLifeTime) * time.Second)

		_log.Info("PostgreSQL写库连接成功")
	} else if c.PostgreSQL != nil {
		// 兼容旧配置
		writeDB, err = gorm.Open(postgres.Open(c.PostgreSQL.Source), gormConfig)
		if err != nil {
			_log.Errorf("连接PostgreSQL写库(兼容模式)失败: %v", err)
			return nil, nil, err
		}

		// 获取原生连接池并设置连接参数
		sqlDB, err := writeDB.DB()
		if err != nil {
			_log.Errorf("获取PostgreSQL写库(兼容模式)原生连接失败: %v", err)
			return nil, nil, err
		}

		sqlDB.SetMaxIdleConns(c.PostgreSQL.MaxIdleConns)
		sqlDB.SetMaxOpenConns(c.PostgreSQL.MaxOpenConns)
		sqlDB.SetConnMaxIdleTime(time.Duration(c.PostgreSQL.ConnMaxIdleTime) * time.Second)
		sqlDB.SetConnMaxLifetime(time.Duration(c.PostgreSQL.ConnMaxLifeTime) * time.Second)

		_log.Info("PostgreSQL写库(兼容模式)连接成功")
	} else {
		_log.Error("PostgreSQL写库配置缺失")
		return nil, nil, err
	}

	// 初始化读库连接
	if c.PostgreSQLRead != nil {
		readDB, err = gorm.Open(postgres.Open(c.PostgreSQLRead.Source), gormConfig)
		if err != nil {
			_log.Errorf("连接PostgreSQL读库失败: %v", err)
			return nil, nil, err
		}

		// 获取原生连接池并设置连接参数
		sqlDB, err := readDB.DB()
		if err != nil {
			_log.Errorf("获取PostgreSQL读库原生连接失败: %v", err)
			return nil, nil, err
		}

		sqlDB.SetMaxIdleConns(c.PostgreSQLRead.MaxIdleConns)
		sqlDB.SetMaxOpenConns(c.PostgreSQLRead.MaxOpenConns)
		sqlDB.SetConnMaxIdleTime(time.Duration(c.PostgreSQLRead.ConnMaxIdleTime) * time.Second)
		sqlDB.SetConnMaxLifetime(time.Duration(c.PostgreSQLRead.ConnMaxLifeTime) * time.Second)

		_log.Info("PostgreSQL读库连接成功")
	} else if writeDB != nil {
		// 如果没有配置读库，使用写库作为读库
		readDB = writeDB
		_log.Info("PostgreSQL读库使用写库配置")
	} else {
		_log.Error("PostgreSQL读库配置缺失")
		return nil, nil, err
	}

	// 清理函数，用于关闭数据库连接
	cleanup := func() {
		_log.Info("关闭PostgreSQL连接...")

		if readDB != nil && readDB != writeDB {
			db, _ := readDB.DB()
			if db != nil {
				_ = db.Close()
			}
		}

		if writeDB != nil {
			db, _ := writeDB.DB()
			if db != nil {
				_ = db.Close()
			}
		}
	}

	return &PostgreSQLClient{
		Read:  readDB,
		Write: writeDB,
		log:   _log,
	}, cleanup, nil
}

// GetDB 获取默认的GORM DB实例（写库）
func (c *PostgreSQLClient) GetDB() *gorm.DB {
	return c.Write
}
