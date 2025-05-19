package providers

import (
	"gil_teacher/app/dao"
	behaviorDao "gil_teacher/app/dao/behavior"
	"gil_teacher/app/dao/live_room/impl"
	"gil_teacher/app/dao/resource_favorite"
	dao_task "gil_teacher/app/dao/task"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// 提供PostgreSQL的GORM DB
func ProvidePostgreSQLDB(client *dao.PostgreSQLClient) *gorm.DB {
	return client.GetDB()
}

// ResourceFavoriteDAOProvider 提供资源收藏DAO
func ResourceFavoriteDAOProvider(db *gorm.DB) *resource_favorite.ResourceFavoriteDAO {
	return resource_favorite.NewResourceFavoriteDAO(db)
}

var RepoProviderSet = wire.NewSet(
	dao.NewActivityDB,           // 提供ActivityDB实例
	dao.NewDBTestClient,         // 提供DBTestClient实例
	dao.NewPostgreSQLClient,     // 提供PostgreSQLClient实例
	dao.NewClickHouseRWClient,   // 提供ClickHouse读写实例
	impl.NewLiveRoomDao,         // 提供LiveRoom DAO
	ProvidePostgreSQLDB,         // 提供GORM DB实例
	dao.NewApiRedisClient,       // 提供API Redis实例
	dao_task.TaskDAOProvider,    // 提供任务数据DAO
	ResourceFavoriteDAOProvider, // 提供资源收藏DAO
	ResourceDAOProvider,         // 提供资源DAO
	FileRecordDAOProvider,       // 提供文件记录DAO
	behaviorDao.NewBehaviorDAO,  // 提供行为DAO
)
