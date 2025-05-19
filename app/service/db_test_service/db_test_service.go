package db_test_service

import (
	"context"

	"gil_teacher/app/conf"
	"gil_teacher/app/dao"

	"github.com/go-kratos/kratos/v2/log"
)

// DBTestService 数据库测试服务
type DBTestService struct {
	log       *log.Helper
	dbTestDao *dao.DBTestClient
	config    *conf.Config
}

// NewDBTestService 创建数据库测试服务
func NewDBTestService(
	logger log.Logger,
	dbTestDao *dao.DBTestClient,
	config *conf.Config,
) *DBTestService {
	return &DBTestService{
		log:       log.NewHelper(log.With(logger, "x_module", "service/NewDBTestService")),
		dbTestDao: dbTestDao,
		config:    config,
	}
}

// TestPostgreSQLConnection 测试PostgreSQL连接
func (s *DBTestService) TestPostgreSQLConnection(ctx context.Context) (bool, error) {
	return s.dbTestDao.TestPostgreSQLConnection()
}

// TestPostgreSQLReadWriteConnections 测试PostgreSQL读写分离连接
func (s *DBTestService) TestPostgreSQLReadWriteConnections(ctx context.Context) map[string]interface{} {
	return s.dbTestDao.TestPostgreSQLReadWriteConnections()
}

// TestClickHouseConnection 测试ClickHouse连接
func (s *DBTestService) TestClickHouseConnection(ctx context.Context) (bool, error) {
	return s.dbTestDao.TestClickHouseConnection()
}

// TestClickHouseWriteConnection 测试ClickHouse写连接
func (s *DBTestService) TestClickHouseWriteConnection(ctx context.Context) (bool, error) {
	return s.dbTestDao.TestClickHouseWriteConnection()
}

// TestClickHouseReadConnection 测试ClickHouse读连接
func (s *DBTestService) TestClickHouseReadConnection(ctx context.Context) (bool, error) {
	return s.dbTestDao.TestClickHouseReadConnection()
}

// TestClickHouseReadWriteConnections 测试ClickHouse读写分离连接
func (s *DBTestService) TestClickHouseReadWriteConnections(ctx context.Context) map[string]interface{} {
	return s.dbTestDao.TestClickHouseReadWriteConnections()
}

// TestAllDBConnections 测试所有数据库连接
func (s *DBTestService) TestAllDBConnections(ctx context.Context) map[string]interface{} {
	return s.dbTestDao.TestAllDBConnections()
}
