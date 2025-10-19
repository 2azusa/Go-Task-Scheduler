package logger

import (
	"fmt"
	"os"
	"pulse/common/pkg/utils"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 全局唯一zap日志实例
var _defaultLogger *zap.Logger

// 初始化全局日志
func Init(projectName string, level string, format, prefix, director string, showLine bool, encodeLevel string, stacktraceKey string, logInConsole bool) (logger *zap.Logger) {
	if ok := utils.Exists(fmt.Sprintf("%s/%s", projectName, director)); !ok {
		fmt.Printf("create %v directory\n", director)
		_ = os.Mkdir(fmt.Sprintf("%s/%s", projectName, director), os.ModePerm)
	}

	debugPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.DebugLevel
	})
	infoPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.InfoLevel
	})
	warnPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev == zap.ErrorLevel
	})
	errorPriority := zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
		return lev >= zap.ErrorLevel
	})

	// 根据日志级别，构建zapcore.Core
	cores := make([]zapcore.Core, 0)
	switch level {
	case "error":
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	case "warn":
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_warn.log", projectName, director), warnPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	case "info":
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_info.log", projectName, director), infoPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_warn.log", projectName, director), warnPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	default:
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_debug.log", projectName, director), debugPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_info.log", projectName, director), infoPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_warn.log", projectName, director), warnPriority))
		cores = append(cores, getEncoderCore(logInConsole, prefix, format, encodeLevel, stacktraceKey, fmt.Sprintf("%s/%s/server_error.log", projectName, director), errorPriority))
	}

	// 合并所有Core并创建logger
	logger = zap.New(zapcore.NewTee(cores[:]...), zap.AddCaller())

	if showLine {
		// 添加调用者信息
		logger = logger.WithOptions(zap.AddCaller())
	}

	_defaultLogger = logger

	return _defaultLogger
}

// Shutdown 优雅关闭，用于确保所有挂起的日志都被写入
func Shutdown() {
	_defaultLogger.Sync()
}

// GetLogger 返回全局的_defaultLogger实例
func GetLogger() *zap.Logger {
	return _defaultLogger
}

// Sync 用于将缓冲区中的日志刷出(Flush)到目的地
func Sync() error {
	return _defaultLogger.Sync()
}

// getEncoderConfig 返回zap的编码设置
func getEncoderConfig(prefix, encodeLevel, stacktraceKey string) (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:    "message",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "name",
		CallerKey:     "caller",
		StacktraceKey: stacktraceKey,
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseColorLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format(prefix + utils.TimeFormatDateV4))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	switch encodeLevel {
	case "LowercaseLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseLevelEncoder // "info"
	case "LowercaseColorLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder // 彩色的 "info"
	case "CapitalLevelEncoder":
		config.EncodeLevel = zapcore.CapitalLevelEncoder // "INFO"
	case "CapitalColorLevelEncoder":
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder // 彩色的 "INFO"
	default:
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return config
}

// getEncoder 返回一个zap编码器
func getEncoder(prefix, format, encodeLevel, stacktraceKey string) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(getEncoderConfig(prefix, encodeLevel, stacktraceKey))
	}
	return zapcore.NewConsoleEncoder(getEncoderConfig(prefix, encodeLevel, stacktraceKey))
}

// getEncoderCore 辅助函数，返回一个完整的zapcore.Core
func getEncoderCore(logInConsole bool, prefix, format, encodeLevel, stacktraceKey string, fileName string, level zapcore.LevelEnabler) (core zapcore.Core) {
	writer := getWriteSyncer(logInConsole, fileName)
	return zapcore.NewCore(getEncoder(prefix, format, encodeLevel, stacktraceKey), writer, level)
}

// getWriteSyncer 返回一个日志写入器
func getWriteSyncer(logInConsole bool, file string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    10,
		MaxBackups: 200,
		MaxAge:     30,
		Compress:   true,
	}

	// 输出到控制台并写入到lumberjack
	if logInConsole {
		return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
	}
	// 写入到lumberjack
	return zapcore.AddSync(lumberJackLogger)
}
