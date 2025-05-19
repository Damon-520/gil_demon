package dao

import (
	"gorm.io/gorm"
)

type AutumnHarvestDB struct {
	AutumnHarvestR *gorm.DB // AutumnHarvestRead
	AutumnHarvestW *gorm.DB // AutumnHarvestWrite
}

// func NewAutumnHarvestDB(c *conf.Data, logger log.Logger) (*AutumnHarvestDB, func(), error) {

// 	_log := log.NewHelper(log.With(logger, "x_module", "data/NewAutumnHarvestDB"))

// 	dbR_ := polardbx.NewPolarDB(c.AutumnHarvestRead, logger)
// 	dbW_ := polardbx.NewPolarDB(c.AutumnHarvestWrite, logger)

// 	cleanup := func() {
// 		log.NewHelper(logger).Info("closing the data resources")

// 		if dbR_ != nil {
// 			db, _ := dbR_.DB()
// 			_ = db.Close()
// 		}
// 		if dbW_ != nil {
// 			db, _ := dbW_.DB()
// 			_ = db.Close()
// 		}

// 		_log.Info("closing the data resources")
// 	}

// 	return &AutumnHarvestDB{
// 		AutumnHarvestR: dbR_,
// 		AutumnHarvestW: dbW_,
// 	}, cleanup, nil

// }
