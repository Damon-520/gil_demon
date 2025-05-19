package gorm_builder

import (
	"fmt"
	"gorm.io/gorm"
)

type Config struct {
	disableBuildWhere bool
	disableBuildOrder bool
	disableBuildSort  bool
	disableBuildOr    bool
}

// type IBuilder interface {
// 	buildWhere()
// 	builderOr()
// 	buildOrder()
// 	builderLimit()
// 	builderOffset()
// 	builderSkip()
// }

type builder struct {
	db     *gorm.DB
	query  *Query
	config *Config
}

func NewBuilder(db *gorm.DB, query *Query) *builder {

	return &builder{
		db:    db,
		query: query,
	}
}

func (q *builder) Config(config *Config) *builder {
	q.config = config
	return q
}

func (q *builder) buildWhere() {
	if q.config.disableBuildWhere == true {
		return
	}
	fmt.Println("builder where....")

	for _, cond := range *q.query.Where {
		// TODO 验证Operator
		sql := fmt.Sprintf("%s %s ?", cond.Field, cond.Operator) // TODO
		q.db = q.db.Where(sql, cond.Value)
	}
}

func (q *builder) buildOr() {
	if q.config.disableBuildOr == true {
		return
	}
	fmt.Println("builder or ....")
}

func (q *builder) buildOrder() {
	if q.config.disableBuildOrder == true {
		return
	}
	fmt.Println("builder order ....")
}

func (q *builder) Builder() *gorm.DB {

	q.buildWhere()
	q.buildOr()
	q.buildOrder()

	return q.db
}
