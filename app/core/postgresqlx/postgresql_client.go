package postgresqlx

import (
	"gil_teacher/app/conf"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type PgDB struct {
	*gorm.DB
}

func NewPgDB(c *conf.Data, logger log.Logger) (*PgDB, func(), error) {

	_log := log.NewHelper(log.With(logger, "x_module", "data/NewPgDB"))

	pgDB := newPostgreSqlDB(c.PostgreSQL, logger)

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")

		if pgDB != nil {
			db, _ := pgDB.DB()
			_ = db.Close()
		}
		_log.Info("closing the data resources")
	}

	return &PgDB{pgDB}, cleanup, nil

}
