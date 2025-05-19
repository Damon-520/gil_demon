package mysqlx

import (
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/core/logger"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

func NewMysqlDB(conf *conf.MySQL, logger_ log.Logger) *gorm.DB {

	log_ := log.NewHelper(log.With(logger_, "x_module", "data/NewMysqlDB"))

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       conf.Source, // DSN data source name
		DefaultStringSize:         256,         // string 类型字段的默认长度
		DisableDatetimePrecision:  true,        // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,        // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,        // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,       // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 logger.NewGorm(logger_),
	})
	if err != nil {
		return nil
		log.Fatalf("mysql TODO")
	}

	db = db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(gormLog.Info),
	})
	err = db.Use(
		dbresolver.Register(dbresolver.Config{Replicas: []gorm.Dialector{mysql.Open(conf.Source)}}).
			SetConnMaxLifetime(time.Hour).
			SetMaxIdleConns(int(conf.MaxIdleConns)).
			SetMaxOpenConns(int(conf.MaxOpenConns)),
	)
	if err != nil {
		log_.Fatalf("failed dbr use to mysql: %v", err)
	}
	return db
}
