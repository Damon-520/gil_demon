package holox

import (
	"fmt"

	"gil_teacher/app/conf"
	"gil_teacher/app/core/logger"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewHoloDB(conf *conf.HoloDB, logger_ log.Logger) *gorm.DB {

	log_ := log.NewHelper(log.With(logger_, "x_module", "core/NewHoloDB"))
	log_.Info("[boot] start NewHoloDB dbr.")

	c := conf
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", c.Username, c.Password, c.Database, c.Host, c.Port)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		Logger: logger.NewGorm(logger_),
	})

	db.Exec(fmt.Sprintf(`set search_path='%s'`, c.Schema))

	if err != nil {
		log_.Fatalf("[boot] failed opening connection to holoDB: %v", err)
	} else {
		log_.Info("[boot] NewHoloDB success.")
	}

	return db
}

// func NewHoloAutumnHarvestRead(conf *conf.Data, logger log.Logger) *gormHahR {
//
// 	log_ := log.NewHelper(log.With(logger, "x_module", "core/NewHoloAutumnHarvestRead"))
// 	log_.Info("[boot] start NewHoloAutumnHarvestRead dbr.")
//
// 	c := conf.HoloAutumnHarvestRead
// 	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", c.Username, c.Password, c.Database, c.Host, c.Port)
// 	db, err := gorm.Open(postgres.New(postgres.Config{
// 		DSN:                  dsn,
// 		PreferSimpleProtocol: true, // disables implicit prepared statement usage
// 	}), &gorm.Config{})
//
// 	db.Exec(fmt.Sprintf(`set search_path='%s'`, c.Schema))
//
// 	if err != nil {
// 		log_.Fatalf("[boot] failed opening connection to NewHoloAutumnHarvestRead: %v", err)
// 	} else {
// 		log_.Info("[boot] NewHoloAutumnHarvestRead success.")
// 	}
//
// 	return (*gormHahR)(db)
// }
//
// func NewHoloAutumnHarvestWrite(conf *conf.Data, logger log.Logger) *gormHahW {
//
// 	log_ := log.NewHelper(log.With(logger, "x_module", "core/NewHoloAutumnHarvestRead"))
// 	log_.Info("[boot] start NewHoloAutumnHarvestWrite dbr.")
//
// 	c := conf.HoloAutumnHarvestWrite
// 	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", c.Username, c.Password, c.Database, c.Host, c.Port)
// 	db, err := gorm.Open(postgres.New(postgres.Config{
// 		DSN:                  dsn,
// 		PreferSimpleProtocol: true, // disables implicit prepared statement usage
// 	}), &gorm.Config{})
//
// 	db.Exec(fmt.Sprintf(`set search_path='%s'`, c.Schema))
//
// 	if err != nil {
// 		log_.Fatalf("[boot] failed opening connection to NewHoloAutumnHarvestWrite: %v", err)
// 	} else {
// 		log_.Info("[boot] NewHoloAutumnHarvestWrite success.")
// 	}
//
// 	return (*gormHahW)(db)
// }
//
// func NewHoloDataCenterRead(conf *conf.Data, logger log.Logger) *gormHdcR {
//
// 	log_ := log.NewHelper(log.With(logger, "x_module", "core/NewHoloDataCenterRead"))
// 	log_.Info("[boot] start NewHoloDataCenterRead dbr.")
//
// 	c := conf.HoloDataCenterRead
// 	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", c.Username, c.Password, c.Database, c.Host, c.Port)
// 	db, err := gorm.Open(postgres.New(postgres.Config{
// 		DSN:                  dsn,
// 		PreferSimpleProtocol: true,
// 	}), &gorm.Config{})
//
// 	db.Exec(fmt.Sprintf(`set search_path='%s'`, c.Schema))
//
// 	if err != nil {
// 		log_.Fatalf("[boot] failed opening connection to NewHoloDataCenterRead: %v", err)
// 	} else {
// 		log_.Info("[boot] NewHoloDataCenterRead success.")
// 	}
//
// 	return (*gormHdcR)(db)
// }
//
// func NewHoloDataCenterWrite(conf *conf.Data, logger log.Logger) *gormHdcW {
//
// 	log_ := log.NewHelper(log.With(logger, "x_module", "core/NewHoloDataCenterWrite"))
// 	log_.Info("[boot] start NewHoloDataCenterWrite dbr.")
//
// 	c := conf.HoloDataCenterWrite
// 	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", c.Username, c.Password, c.Database, c.Host, c.Port)
// 	db, err := gorm.Open(postgres.New(postgres.Config{
// 		DSN:                  dsn,
// 		PreferSimpleProtocol: true, // disables implicit prepared statement usage
// 	}), &gorm.Config{})
//
// 	db.Exec(fmt.Sprintf(`set search_path='%s'`, c.Schema))
//
// 	if err != nil {
// 		log_.Fatalf("[boot] failed opening connection to NewHoloDataCenterWrite: %v", err)
// 	} else {
// 		log_.Info("[boot] NewHoloDataCenterWrite success.")
// 	}
//
// 	return (*gormHdcW)(db)
// }
