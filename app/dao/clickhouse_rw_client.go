package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gil_teacher/app/conf"
	"gil_teacher/app/consts"
	"gil_teacher/app/core/logger"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Model 基础模型接口
type Model interface {
	// 获取表名
	TableName() string
	// 生成 ID
	GenerateID(ctx context.Context) string
}

// ClickHouseRWClient 定义 ClickHouse 客户端结构体
type ClickHouseRWClient struct {
	writeClient driver.Conn
	readClient  driver.Conn
	model       Model
	log         *logger.ContextLogger
}

// NewClickHouseRWClient 创建 ClickHouse 读写分离客户端
func NewClickHouseRWClient(c *conf.Data, logger *logger.ContextLogger) (map[string]*ClickHouseRWClient, func(), error) {
	ctx := context.Background()
	multiWriteClients, err := connectToClickHouse(c.ClickhouseWrite)
	if err != nil {
		logger.Error(ctx, "failed to connect to write ClickHouse, config: %+v, error: %v", c.ClickhouseWrite, err)
		return nil, nil, fmt.Errorf("failed to connect to write ClickHouse: %w", err)
	}

	multiReadClients, err := connectToClickHouse(c.ClickhouseRead)
	if err != nil {
		logger.Error(ctx, "failed to connect to read ClickHouse, config: %+v, error: %v", c.ClickhouseRead, err)
		return nil, nil, fmt.Errorf("failed to connect to read ClickHouse: %w", err)
	}

	// 清理函数，用于关闭数据库连接
	cleanup := func() {
		logger.Info(ctx, "关闭ClickHouse连接...")
		for _, writeClient := range multiWriteClients {
			if err := writeClient.Close(); err != nil {
				logger.Error(ctx, "failed to close write client: %v", err)
			}
		}
		for _, readClient := range multiReadClients {
			if err := readClient.Close(); err != nil {
				logger.Error(ctx, "failed to close read client: %v", err)
			}
		}
	}

	clients := make(map[string]*ClickHouseRWClient, len(multiWriteClients))
	for dbName, writeClient := range multiWriteClients {
		clients[dbName] = &ClickHouseRWClient{
			writeClient: writeClient,
			readClient:  multiReadClients[dbName],
			log:         logger,
		}
	}

	return clients, cleanup, nil
}

func connectToClickHouse(config *conf.Clickhouse) (map[string]driver.Conn, error) {
	if len(config.Databases) > 0 {
		conns := make(map[string]driver.Conn, 0)
		for _, dbName := range config.Databases {
			conn, err := connect(config, dbName)
			if err != nil {
				return nil, err
			}
			conns[dbName] = conn
		}
		return conns, nil
	} else {
		conn, err := connect(config, config.Database)
		if err != nil {
			return nil, err
		}
		return map[string]driver.Conn{config.Database: conn}, nil
	}
}

// connectToClickHouse 根据配置连接到 ClickHouse
func connect(config *conf.Clickhouse, database string) (driver.Conn, error) {
	opts := &clickhouse.Options{
		Addr: config.Address,
		Auth: clickhouse.Auth{
			Database: database,
			Username: config.Username,
			Password: config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": config.MaxExecutionTime,
		},
		ReadTimeout: time.Second * 10,
		DialTimeout: time.Duration(config.DialTimeout) * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	}

	// 启用调试日志
	opts.Debug = true
	opts.Debugf = func(format string, v ...any) {
		fmt.Printf("DEBUG: "+format+"\n", v...)
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open ClickHouse: %w", err)
	}
	if err = conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("conn failed to ping ClickHouse: %w", err)
	}

	return conn, nil
}

func (c *ClickHouseRWClient) Model(dest any) *ClickHouseRWClient {
	c.model = dest.(Model)
	return c
}

// Prepare 准备执行写操作
func (c *ClickHouseRWClient) Prepare(ctx context.Context, query string) (driver.Batch, error) {
	return c.writeClient.PrepareBatch(ctx, query)
}

// Exec 执行写操作
func (c *ClickHouseRWClient) Exec(ctx context.Context, query string, args ...any) error {
	batch, err := c.writeClient.PrepareBatch(ctx, query)
	if err != nil {
		return err
	}
	batch.Append(args...)
	err = batch.Send()
	if err != nil {
		return err
	}
	return nil
}

// PrepareExecModel 准备并执行写操作，args 为结构体
func (c *ClickHouseRWClient) PrepareExecModel(ctx context.Context, query string, models any) error {
	batch, err := c.writeClient.PrepareBatch(ctx, query)
	if err != nil {
		return err
	}
	// models 可能是切片，需要遍历
	modelsValue := reflect.ValueOf(models)
	if modelsValue.Kind() == reflect.Slice {
		for i := range modelsValue.Len() {
			batch.AppendStruct(modelsValue.Index(i).Interface())
		}
	} else {
		batch.AppendStruct(models)
	}
	err = batch.Send()
	if err != nil {
		return err
	}
	return nil
}

// Read 执行读操作，支持批量结果查询和数量查询
// dest 必须是指针类型，且可能是切片指针结构，或者切片结构，即 dest 的类型为 *[]T 或 *[]*T
// 如果 dest 是切片，则返回切片，否则返回单个结构体
// 如果 dest 是 int64 类型，则用于查询数量
func (c *ClickHouseRWClient) Read(ctx context.Context, dest any, query string, args ...any) error {
	rows, err := c.readClient.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// 获取 dest 的类型
	destType := reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	// 获取指针指向的实际类型
	elemType := destType.Elem()

	// 处理数量查询
	if elemType.Kind() == reflect.Uint64 {
		if !rows.Next() {
			return errors.New("no rows returned")
		}
		if err := rows.Scan(dest); err != nil {
			return fmt.Errorf("scan count failed: %w", err)
		}
		return nil
	}

	// 处理切片类型
	if elemType.Kind() == reflect.Slice {
		// 创建切片
		sliceValue := reflect.MakeSlice(elemType, 0, 0)
		// 获取切片元素类型
		elemType = elemType.Elem()

		// 处理切片元素是指针的情况
		isPointerElem := elemType.Kind() == reflect.Ptr
		if isPointerElem {
			elemType = elemType.Elem()
		}

		// 确保元素类型是结构体
		if elemType.Kind() != reflect.Struct {
			return fmt.Errorf("slice element must be a struct, got %v", elemType.Kind())
		}

		// 遍历结果集
		for rows.Next() {
			// 创建新的元素
			var newElem reflect.Value
			if isPointerElem {
				newElem = reflect.New(elemType)
			} else {
				newElem = reflect.New(elemType).Elem()
			}

			// 扫描数据到新元素
			if err := rows.ScanStruct(newElem.Interface()); err != nil {
				return fmt.Errorf("scan struct to slice failed: %w", err)
			}

			// 追加到切片中
			if isPointerElem {
				sliceValue = reflect.Append(sliceValue, newElem)
			} else {
				sliceValue = reflect.Append(sliceValue, newElem)
			}
		}

		// 设置结果到目标切片
		reflect.ValueOf(dest).Elem().Set(sliceValue)
	} else {
		// 确保目标类型是结构体
		if elemType.Kind() != reflect.Struct {
			return fmt.Errorf("dest must be a pointer to struct, got pointer to %v", elemType.Kind())
		}

		// 处理单个结构体
		if !rows.Next() {
			return errors.New("no rows returned")
		}
		if err := rows.ScanStruct(dest); err != nil {
			return fmt.Errorf("scan single struct failed: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration failed: %w", err)
	}

	return nil
}

// Close 关闭客户端连接
func (c *ClickHouseRWClient) Close() error {
	if err := c.writeClient.Close(); err != nil {
		return fmt.Errorf("failed to close write client: %w", err)
	}
	if err := c.readClient.Close(); err != nil {
		return fmt.Errorf("failed to close read client: %w", err)
	}
	return nil
}

// Find 查询单条记录
func (db *ClickHouseRWClient) Find(ctx context.Context, dest any, where map[string]any) error {
	if len(where) == 0 {
		return errors.New("where conditions cannot be empty")
	}

	conditions, args := db.buildWhere(where)
	query := "SELECT * FROM " + db.model.TableName() + " WHERE " + strings.Join(conditions, " AND ") + " LIMIT 1"
	err := db.Read(ctx, dest, query, args...)
	if err != nil {
		db.log.Error(ctx, "Find query failed: %v", err)
		return err
	}

	return nil
}

// FindAll 查询多条记录
func (db *ClickHouseRWClient) FindAll(ctx context.Context, dest any, where map[string]any, pageInfo *consts.DBPageInfo) error {
	return db.FindAllWithFields(ctx, dest, nil, where, pageInfo)
}

// 查询多条记录，并指定返回字段
func (db *ClickHouseRWClient) FindAllWithFields(ctx context.Context, dest any, fields []string, where map[string]any, pageInfo *consts.DBPageInfo) error {
	if len(fields) == 0 {
		fields = []string{"*"}
	}

	// 构建查询语句
	conditions, args := db.buildWhere(where)
	query := "SELECT " + strings.Join(fields, ",") + " FROM " + db.model.TableName()
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// 分页和排序
	pageInfo = consts.DefaultDBPageInfo(pageInfo)
	if pageInfo.SortBy != "" {
		query += fmt.Sprintf(" ORDER BY %s %s", pageInfo.SortBy, pageInfo.SortType)
	}
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageInfo.Limit, (pageInfo.Page-1)*pageInfo.Limit)
	err := db.Read(ctx, dest, query, args...)
	if err != nil {
		db.log.Error(ctx, "查询失败: %v", err)
		return err
	}

	return nil
}

// 总数
func (db *ClickHouseRWClient) Count(ctx context.Context, where map[string]any) (int64, error) {
	query := "SELECT COUNT(*) FROM " + db.model.TableName()
	conditions, args := db.buildWhere(where)
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var count uint64
	err := db.Read(ctx, &count, query, args...)
	if err != nil {
		db.log.Error(ctx, "查询数量失败: %v", err)
		return 0, err
	}

	return int64(count), nil
}

// Insert 插入单条记录，并返回插入的 ID
func (db *ClickHouseRWClient) Insert(ctx context.Context, value any) (string, error) {
	uuid := value.(Model).GenerateID(ctx)
	batch, err := db.Prepare(ctx, "INSERT INTO "+db.model.TableName())
	if err != nil {
		db.log.Error(ctx, "准备批处理失败, value: %+v, err: %v", value, err)
		return "", err
	}

	if err := batch.AppendStruct(value); err != nil {
		db.log.Error(ctx, "添加数据失败, value: %+v, err: %v", value, err)
		return "", err
	}

	err = batch.Send()
	if err != nil {
		db.log.Error(ctx, "发送数据失败, value: %+v, err: %v", value, err)
		return "", err
	}
	return uuid, nil
}

// BatchInsert 批量插入记录
// values 必须是切片类型，且切片元素必须实现 Model 接口
func (db *ClickHouseRWClient) BatchInsert(ctx context.Context, values any) ([]string, error) {
	// 获取 values 的类型
	valuesType := reflect.TypeOf(values)
	if valuesType == nil {
		return nil, errors.New("values cannot be nil")
	}

	// 确保 values 是切片类型
	if valuesType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("values must be a slice, got %v", valuesType.Kind())
	}

	// 获取切片值
	valuesValue := reflect.ValueOf(values)
	if valuesValue.Len() == 0 {
		return nil, nil
	}

	// 获取切片元素类型
	// elemType := valuesType.Elem()
	// // 确保元素类型实现了 Model 接口
	// if !reflect.PointerTo(elemType).Implements(reflect.TypeOf((*Model)(nil)).Elem()) {
	// 	return nil, fmt.Errorf("slice elements must implement Model interface, got %v", elemType)
	// }

	// 生成 UUID
	uuids := make([]string, 0, valuesValue.Len())
	for i := range valuesValue.Len() {
		elem := valuesValue.Index(i).Interface()
		uuids = append(uuids, elem.(Model).GenerateID(ctx))
	}

	var lastErr error
	for i := range consts.ChMaxInsertRetry {
		query := fmt.Sprintf("INSERT INTO %s ", db.model.TableName())
		success := true

		err := db.PrepareExecModel(ctx, query, values)
		// 超时或者连接断开重试
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, sql.ErrConnDone) {
				lastErr = err
				success = false
				db.log.Error(ctx, "执行批量插入失败，重试次数 %d: %v", i, err)
				time.Sleep(time.Second * time.Duration(i+1))
				continue
			}
			return nil, err
		}

		if success {
			return uuids, nil
		}

		time.Sleep(time.Second * time.Duration(i+1))
	}

	return nil, lastErr
}

// Update 更新记录，并返回影响的行数
func (db *ClickHouseRWClient) Update(ctx context.Context, set map[string]any, where map[string]any) error {
	if len(set) == 0 {
		return errors.New("set conditions cannot be empty")
	}

	// 构建 SET 子句
	setValues := make([]string, 0, len(set))
	setArgs := make([]any, 0, len(set))
	for k, v := range set {
		setValues = append(setValues, fmt.Sprintf("%s = ?", k))
		setArgs = append(setArgs, v)
	}

	// 构建 WHERE 子句
	conditions, whereArgs := db.buildWhere(where)
	query := fmt.Sprintf("ALTER TABLE %s UPDATE %s",
		db.model.TableName(),
		strings.Join(setValues, ", "))

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	args := append(setArgs, whereArgs...)
	// 执行更新操作
	err := db.Exec(ctx, query, args...)
	if err != nil {
		db.log.Error(ctx, "更新数据失败: %v", err)
		return err
	}

	return nil
}

// build where conditions, any 可能是 slice
func (db *ClickHouseRWClient) buildWhere(where map[string]any) ([]string, []any) {
	conditions := make([]string, 0, len(where))
	args := make([]any, 0, len(where))
	for k, v := range where {
		// 如果 v 是 slice，则转换为 IN 查询
		if reflect.TypeOf(v).Kind() == reflect.Slice {
			conditions = append(conditions, k+" IN (?)")
			args = append(args, v)
		} else {
			conditions = append(conditions, k+" = ?")
			args = append(args, v)
		}
	}
	return conditions, args
}

// CHGroupCountResult 用于存储 group by 计数结果
// GroupValues: group by 字段名和值
// Count: 分组计数
type CHGroupCountResult struct {
	GroupValues map[string]interface{} // group by 字段名和值
	Count       int64                  // 计数
}

// CountGroupBy 支持 group by 的计数查询
// groupByFields: 需要 group by 的字段列表
// where: 查询条件
// 返回每个分组的字段值和计数，格式为 []CHGroupCountResult
func (db *ClickHouseRWClient) CountGroupBy(ctx context.Context, groupByFields []string, where map[string]any) ([]CHGroupCountResult, error) {
	if len(groupByFields) == 0 {
		return nil, errors.New("groupByFields cannot be empty")
	}

	// 构建查询语句
	fields := append(groupByFields, "COUNT(*) as cnt")
	conditions, args := db.buildWhere(where)
	query := "SELECT " + strings.Join(fields, ", ") + " FROM " + db.model.TableName()
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " GROUP BY " + strings.Join(groupByFields, ", ")

	// 通过反射获取 model 字段类型
	modelType := reflect.TypeOf(db.model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	rows, err := db.readClient.Query(ctx, query, args...)
	if err != nil {
		db.log.Error(ctx, "CountGroupBy query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	results := make([]CHGroupCountResult, 0)
	for rows.Next() {
		scanArgs := make([]interface{}, len(groupByFields)+1)
		groupVals := make([]interface{}, len(groupByFields))
		for i, field := range groupByFields {
			var ptr interface{}
			var found bool
			// 优先用 ch tag 匹配
			for j := 0; j < modelType.NumField(); j++ {
				f := modelType.Field(j)
				if f.Tag.Get("ch") == field || f.Name == field {
					switch f.Type.Kind() {
					case reflect.Uint64:
						ptr = new(uint64)
					case reflect.String:
						ptr = new(string)
					case reflect.Int64:
						ptr = new(int64)
					case reflect.Float64:
						ptr = new(float64)
					case reflect.Bool:
						ptr = new(bool)
					default:
						ptr = new(interface{})
					}
					found = true
					break
				}
			}
			if !found {
				ptr = new(interface{})
			}
			groupVals[i] = ptr
			scanArgs[i] = ptr
		}
		var cnt uint64
		scanArgs[len(groupByFields)] = &cnt

		if err := rows.Scan(scanArgs...); err != nil {
			db.log.Error(ctx, "CountGroupBy scan failed: %v", err)
			return nil, err
		}

		groupMap := make(map[string]interface{}, len(groupByFields))
		for i, field := range groupByFields {
			switch v := groupVals[i].(type) {
			case *uint64:
				groupMap[field] = *v
			case *string:
				groupMap[field] = *v
			case *int64:
				groupMap[field] = *v
			case *float64:
				groupMap[field] = *v
			case *bool:
				groupMap[field] = *v
			default:
				groupMap[field] = *(v.(*interface{}))
			}
		}
		results = append(results, CHGroupCountResult{
			GroupValues: groupMap,
			Count:       int64(cnt),
		})
	}

	if err := rows.Err(); err != nil {
		db.log.Error(ctx, "CountGroupBy rows iteration failed: %v", err)
		return nil, err
	}

	return results, nil
}
