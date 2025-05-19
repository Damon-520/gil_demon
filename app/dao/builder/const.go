package builder

var (
	Select         = "select"
	SelectDistinct = "distinct"

	GroupBy = "group by"
	OrderBy = "order by"
	Limit   = "limit"
	Offset  = "offset"

	JoinAnd     = "and"
	JoinOr      = "or"
	JoinBetween = "between"

	WhereInSql         = "in"
	WhereIn            = "in"
	WhereNotInSql      = "not in"
	WhereNotIn         = "nIn"
	WhereLikeSql       = "like"
	WhereLike          = "like"
	WhereNotLikeSql    = "not like"
	WhereNotLike       = "nLike"
	WhereIsNullSql     = "is null"
	WhereIsNull        = "null"
	WhereIsNotNullSql  = "is not null"
	WhereIsNotNull     = "nNull"
	WhereBetweenSql    = "between"
	WhereBetween       = "between"
	WhereNotBetweenSql = "not between"
	WhereNotBetween    = "nBetween"
	WhereExistsSql     = "ex"
	WhereExists        = "exists"
	WhereNotExistsSql  = "nEx"
	WhereNotExists     = "not_exists"
	WhereGtSql         = ">"
	WhereGt            = "gt"
	WhereGteSql        = ">="
	WhereGte           = "gte"
	WhereLtSql         = "<"
	WhereLt            = "lt"
	WhereLteSql        = "<="
	WhereLte           = "lte"
	WhereEqSql         = "="
	WhereEq            = "eq"
	WhereNeqSql        = "<>"
	WhereNeq           = "neq"
	WhereRegexpSql     = "regexp"
	WhereRegexp        = "regexp"
	WhereNotRegexpSql  = "not regexp"
	WhereNotRegexp     = "nRegexp"

	whereMap = map[string]string{
		WhereIn:         WhereInSql,
		WhereNotIn:      WhereNotInSql,
		WhereLike:       WhereLikeSql,
		WhereNotLike:    WhereNotLikeSql,
		WhereIsNull:     WhereIsNullSql,
		WhereIsNotNull:  WhereIsNotNullSql,
		WhereBetween:    WhereBetweenSql,
		WhereNotBetween: WhereNotBetweenSql,
		WhereExists:     WhereExistsSql,
		WhereNotExists:  WhereNotExistsSql,
		WhereGt:         WhereGtSql,
		WhereGte:        WhereGteSql,
		WhereLt:         WhereLtSql,
		WhereLte:        WhereLteSql,
		WhereEq:         WhereEqSql,
		WhereNeq:        WhereNeqSql,
		WhereRegexp:     WhereRegexpSql,
		WhereNotRegexp:  WhereNotRegexpSql,
	}

	GroupCount = "count"
	GroupSum   = "sum"
	GroupMax   = "max"
	GroupMin   = "min"
	GroupAvg   = "avg"
)
