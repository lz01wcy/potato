package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"time"
)

func init() {
	initDefaultLogger() // 如果不自定义日志配置 就使用默认初始化 log等级为debug 输出到stdout
}

var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
}

// 只能输出结构化日志，但是性能要高于 SugaredLogger
var Logger *zap.Logger

// 可以输出 结构化日志、非结构化日志
var Sugar *zap.SugaredLogger

var customLevel zap.AtomicLevel

var defaultZapConfig = zapcore.EncoderConfig{
	MessageKey:   "msg",                       //结构化（json）输出：msg的key
	LevelKey:     "level",                     //结构化（json）输出：日志级别的key（INFO，WARN，ERROR等）
	TimeKey:      "time",                      //结构化（json）输出：时间的key（INFO，WARN，ERROR等）
	CallerKey:    "caller",                    //结构化（json）输出：打印日志的文件对应的Key
	EncodeLevel:  zapcore.CapitalLevelEncoder, //将日志级别转换成大写（INFO，WARN，ERROR等）
	EncodeCaller: zapcore.ShortCallerEncoder,  //采用短文件路径编码输出（test/main.go:14 ）
	EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}, //输出的时间格式``
	EncodeDuration: func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendInt64(int64(d) / 1000000)
	},
}

func initDefaultLogger() {
	lv := zap.NewAtomicLevelAt(zapcore.DebugLevel)
	Logger = zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(defaultZapConfig), zapcore.AddSync(os.Stdout), lv), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	Sugar = Logger.Sugar()
}

// InitLogger 初始化日志
// logDir 日志文件夹
// logFile 日志文件名
// logLevel 日志级别
// logMaxSize 每个日志文件保存的最大尺寸 单位:M
// logMaxBackups 日志文件最多保存多少个备份
// logMaxAge 日志文件最多保存多少天
// logCompress 是否压缩
// errorDump 是否单独输出错误日志到文件
func InitLogger(logDir, logFile, logLevel string, logMaxSize, logMaxBackups, logMaxAge int, logCompress, errorDump bool) {
	// 先创建文件夹
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	//自定义日志级别：自定义级别
	l, ok := levelMap[logLevel]
	if !ok {
		println("InitLog error: unknown log level: ", logLevel, " available:[debug|info|warn|error] (defaulting to info)")
		l = zapcore.InfoLevel
	}
	customLevel = zap.NewAtomicLevelAt(l)
	errorLevel := zap.NewAtomicLevelAt(zap.ErrorLevel)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig) // NewConsoleEncoder 是非结构化输出 NewJSONEncoder 是结构化输出

	// 实现多个输出
	// 获取io.Writer
	infoWriter := getWriter(fmt.Sprintf("%s/%s.log", logDir, logFile), logMaxSize, logMaxBackups, logMaxAge, logCompress)
	cores := []zapcore.Core{
		zapcore.NewCore(encoder, zapcore.AddSync(infoWriter), customLevel), //将自定义等级及以上写入正常日志
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), customLevel),  //标准输出
	}
	if errorDump {
		errWriter := getWriter(fmt.Sprintf("%s/%s_err.log", logDir, logFile), logMaxSize, logMaxBackups, logMaxAge, logCompress)
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(errWriter), errorLevel)) //error及以上写入错误日志
	}
	tee := zapcore.NewTee(cores...)

	Logger = zap.New(tee, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	Sugar = Logger.Sugar()
}

func getWriter(filename string, maxSize, maxBackups, maxAge int, compress bool) io.Writer {
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,    //最大M数，超过则切割
		MaxBackups: maxBackups, //最大文件保留数，超过就删除最老的日志文件
		MaxAge:     maxAge,     //保存天数
		Compress:   compress,   //是否压缩
	}
}
