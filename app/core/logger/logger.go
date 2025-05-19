package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

var _ log.Logger = (*Logger)(nil)
var _ logrus.Formatter = (*logFormatter)(nil)

type Option func(logger *Logger)

type Config struct {
	Path         string
	Level        string
	RotationTime time.Duration
	MaxAge       time.Duration
}

type Logger struct {
	logger *logrus.Logger
}

func NewLogger(config Config, ops ...Option) *Logger {
	l := &Logger{}
	l.logger = logrus.New()
	l.logger.SetFormatter(&logFormatter{})

	for _, o := range ops {
		o(l)
	}

	if level, err := logrus.ParseLevel(config.Level); err != nil {
		panic(err.Error())
	} else {
		l.logger.SetLevel(level)
	}

	// 创建多个输出
	var outputs []io.Writer

	// 总是添加标准输出
	outputs = append(outputs, os.Stdout)

	// 如果配置了文件路径，添加文件输出
	if config.Path != "" {
		// 设置日志滚动更新
		writer, err := rotatelogs.New(
			config.Path,
			rotatelogs.WithRotationTime(config.RotationTime),
			rotatelogs.WithMaxAge(config.MaxAge),
		)
		if err != nil {
			panic(err.Error())
		}
		outputs = append(outputs, writer)
	}

	// 使用 MultiWriter 组合多个输出
	multiWriter := io.MultiWriter(outputs...)
	l.logger.SetOutput(multiWriter)

	// 设置默认 msg 字段
	log.DefaultMessageKey = "msg"
	return l
}

func (lf *Logger) Log(level log.Level, keyvals ...interface{}) error {
	buf := make(map[string]interface{})
	for i := 0; i < len(keyvals); i += 2 {
		if logKey := keyvals[i].(string); logKey != "" {
			buf[logKey] = keyvals[i+1]
		}
	}

	switch level {
	case log.LevelFatal:
		lf.fatalM(buf)
	case log.LevelWarn:
		lf.warnM(buf)
	case log.LevelError:
		lf.errorM(buf)
	case log.LevelDebug:
		lf.debugM(buf)
	default:
		lf.infoM(buf)
	}

	return nil
}

type logFormatter struct {
}

func (lf *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields, len(entry.Data)+6)
	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	data["timestamp"] = time.Now().Format("2006-01-02T15:04:05-07:00")
	data["level"] = entry.Level.String()
	hostname, _ := os.Hostname()
	data["host"] = hostname

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON: %v", err)
	}

	return b.Bytes(), nil
}

func (l *Logger) debugM(msgs map[string]interface{}) {
	l.logger.WithFields(msgs).Debug()
}

func (l *Logger) infoM(msgs map[string]interface{}) {
	l.logger.WithFields(msgs).Info()
}

func (l *Logger) warnM(msgs map[string]interface{}) {
	l.logger.WithFields(msgs).Warn()
}

func (l *Logger) errorM(msgs map[string]interface{}) {
	l.logger.WithFields(msgs).Error()
}

func (l *Logger) fatalM(msgs map[string]interface{}) {
	l.logger.WithFields(msgs).Fatal()
}
