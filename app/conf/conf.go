package conf

import (
	"time"
)

// 数据库配置默认值
const (
	// PG_CREATE_BATCH_SIZE_DEFAULT 默认批量创建记录的大小
	PG_CREATE_BATCH_SIZE_DEFAULT = 100
)

type Conf struct {
	App           *App           `json:"app"`
	Server        *Server        `json:"server"`
	Log           *Log           `json:"log"`
	Config        *Config        `json:"config"`
	Redis         *Redis         `json:"redis"`
	Data          *Data          `json:"data"`
	Elasticsearch *Elasticsearch `json:"elasticsearch"`
	ZipKin        *ZipKin        `json:"zipkin"`
	QuestionAPI   *QuestionAPI   `json:"question_api"`
	VolcAI        *VolcAI        `json:"volc_ai"`
}

type QuestionAPI struct {
	Host string `json:"host"`
}

type ZipKin struct {
	Url string `json:"url"`
}

type App struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Mode    int    `json:"mode"`
}

type Server struct {
	Http *Http `json:"http"`
	Grpc *Grpc `json:"grpc"`
}

type Http struct {
	Network string `json:"network"`
	Addr    string `json:"addr"`
	Timeout string `json:"timeout"`
}

type Grpc struct {
	Network string `json:"network"`
	Addr    string `json:"addr"`
	Timeout string `json:"timeout"`
}

type Log struct {
	Path         string `json:"path"`
	Level        string `json:"level"`
	RotationTime string `json:"rotation_time"`
	MaxAge       string `json:"max_age"`
}

type Config struct {
	Env         string        `json:"env"`
	AdminAuth   *AdminAuth    `json:"admin_auth"`
	OSS         *OSSConfig    `json:"oss"`
	Upload      *UploadConfig `json:"upload"`
	GilAdminAPI *GilAdminAPI  `json:"gil_admin_api"`
}

type OSSConfig struct {
	AccessKeyID     string `json:"accessKeyID" yaml:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret" yaml:"accessKeySecret"`
	Region          string `json:"region" yaml:"region"`
	BucketName      string `json:"bucketName" yaml:"bucketName"`
	Endpoint        string `json:"endpoint" yaml:"endpoint"`
	Internal        bool   `json:"internal" yaml:"internal"`
	Secure          bool   `json:"secure" yaml:"secure"`
	BasePath        string `json:"basePath" yaml:"basePath"`
}

type AdminAuth struct {
	Domain   string `json:"domain"`
	SystemId int    `json:"system_id"`
	Timeout  string `json:"timeout"`
}

type Redis struct {
	Network      string `json:"network"`
	Address      string `json:"address"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Database     int    `json:"database"`
	DialTimeout  string `json:"dial_timeout"`
	ReadTimeout  string `json:"read_timeout"`
	WriteTimeout string `json:"write_timeout"`
}

type PolarDB struct {
	Driver       string `json:"driver"`
	Source       string `json:"source"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
}

type HoloDB struct {
	Host     string `json:"host"`
	Port     int32  `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	Schema   string `json:"schema"`
}

type Data struct {
	ActivityWrite   *MySQL      `json:"activity_write"`
	ActivityRead    *MySQL      `json:"activity_read"`
	PostgreSQLWrite *PostgreSQL `json:"postgresql_write"`
	PostgreSQLRead  *PostgreSQL `json:"postgresql_read"`
	PolarDB         *PolarDB    `json:"polar_db"`
	Redis           *Redis      `json:"redis"`
	PostgreSQL      *PostgreSQL `json:"postgresql"`
	Clickhouse      *Clickhouse `json:"clickhouse"`
	ClickhouseWrite *Clickhouse `json:"clickhouse_write"`
	ClickhouseRead  *Clickhouse `json:"clickhouse_read"`
	Mongo           *Mongo      `json:"mongo"`
	Kafka           *Kafka      `json:"kafka"`
	Nacos           *Nacos      `json:"nacos"`
}

type MySQL struct {
	Driver       string `json:"driver"`
	Source       string `json:"source"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
}

type PostgreSQL struct {
	Driver          string `json:"driver"`
	Source          string `json:"source"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	MaxOpenConns    int    `json:"max_open_conns"`
	ConnMaxIdleTime int    `json:"conn_max_idle_time"`
	ConnMaxLifeTime int    `json:"conn_max_lifetime"`
	ConnectTimeout  int    `json:"connect_timeout"`
	CreateBatchSize int    `json:"create_batch_size"`
}

type Clickhouse struct {
	Address          []string `json:"address"`
	Database         string   `json:"database"`
	Databases        []string `json:"databases"`
	Username         string   `json:"username"`
	Password         string   `json:"password"`
	MaxExecutionTime int      `json:"max_execution_time"`
	DialTimeout      int      `json:"dial_timeout"`
	ReadTimeout      int      `json:"read_timeout"`
}

type Mongo struct {
	Url string `json:"url,omitempty"`
}

// ProducerConfig 生产者配置
type ProducerConfig struct {
	RetryMax int           `json:"retryMax"`
	Timeout  time.Duration `json:"timeout"`
}

// ConsumerConfig 消费者配置
type ConsumerConfig struct {
	AutoCommit     bool          `json:"autoCommit"`
	CommitInterval time.Duration `json:"commitInterval"`
	MaxPollRecords int           `json:"maxPollRecords"`
	BatchSize      int           `json:"batchSize"`
	BatchTime      time.Duration `json:"batchTime"`
	SessionTime    time.Duration `json:"sessionTime"`
}

type Kafka struct {
	Version  string         `json:"version"`
	Brokers  string         `json:"broker"`
	Producer ProducerConfig `json:"producer"`
	Consumer ConsumerConfig `json:"consumer"`
}

type Elasticsearch struct {
	EsURL    string `json:"esurl"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type GilAdminAPI struct {
	UcenterHost string `json:"ucenter_host"`
	AdminHost   string `json:"admin_host"`
}

// Bootstrap 配置引导程序
type Bootstrap struct {
	Upload *UploadConfig `json:"upload"`
}

// UploadConfig 上传相关配置
type UploadConfig struct {
	Callback *CallbackConfig `json:"callback"`
}

// CallbackConfig 回调相关配置
type CallbackConfig struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

// NewBootstrap 创建配置引导程序
func NewBootstrap(c *Conf) *Bootstrap {
	return &Bootstrap{
		Upload: c.Config.Upload,
	}
}

type Nacos struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// VolcAI 火山引擎AI配置
type VolcAI struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}
