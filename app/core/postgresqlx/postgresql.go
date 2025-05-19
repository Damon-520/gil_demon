package postgresqlx

import (
	"fmt"
	"time"

	c "gil_teacher/app/conf"
	"gil_teacher/app/core/logger"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLog "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

func newPostgreSqlDB(conf *c.PostgreSQL, logger_ log.Logger) *gorm.DB {

	log_ := log.NewHelper(log.With(logger_, "x_module", "data/PostgreSqlDB"))

	//追加配置参数
	dsn := fmt.Sprintf(
		"%s connect_timeout=%d ",
		conf.Source,
		conf.ConnectTimeout,
	)

	// 设置批量创建的默认大小
	batchSize := conf.CreateBatchSize
	if batchSize <= 0 {
		batchSize = c.PG_CREATE_BATCH_SIZE_DEFAULT
	}

	//初始化主库连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true, // 开启预处理缓存
		SkipDefaultTransaction: true, // 禁用默认事务
		Logger:                 logger.NewGorm(logger_),
		CreateBatchSize:        conf.CreateBatchSize,
	})

	if err != nil {
		log.Fatalf("connect db error:" + err.Error())
	}

	db = db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(gormLog.Info),
	})

	err = db.Use(
		dbresolver.Register(dbresolver.Config{Replicas: []gorm.Dialector{postgres.Open(dsn)}}).
			SetMaxIdleConns(conf.MaxIdleConns).
			SetMaxOpenConns(conf.MaxOpenConns).
			SetConnMaxLifetime(time.Duration(conf.ConnMaxLifeTime) * time.Second).
			SetConnMaxIdleTime(time.Duration(conf.ConnMaxIdleTime) * time.Second),
	)
	if err != nil {
		log_.Fatalf("failed dbr use to postgresqlx: %v", err)
	}

	return db
}
