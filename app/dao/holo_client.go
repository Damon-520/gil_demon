package dao

import (
	"gorm.io/gorm"
)

type HoloDataCenterDB gorm.DB

// func NewHoloDataCenterDB(c *conf.Data, logger log.Logger) (*HoloDataCenterDB, func(), error) {

// 	_log := log.NewHelper(log.With(logger, "x_module", "data/NewHoloDataCenterDB"))

// 	db_ := holox.NewHoloDB(c.HoloDataCenter, logger)

// 	cleanup := func() {
// 		log.NewHelper(logger).Info("closing the data resources")

// 		if db_ != nil {
// 			db, _ := db_.DB()
// 			_ = db.Close()
// 		}

// 		_log.Info("closing the data resources")
// 	}

// 	return (*HoloDataCenterDB)(db_), cleanup, nil

// }
