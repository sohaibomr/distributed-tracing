package logger

// zap logger
import (
	"go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
	"go.elastic.co/apm/module/apmzap/v2"
)

var Log = newLogger()

func newLogger() *zap.Logger {
	// zap logger config
	// config := zap.NewProductionConfig()
	// config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// config.EncoderConfig.TimeKey = "timestamp"
	// config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	// config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	// logger, err := config.Build()
	// if err != nil {
	// 	panic(err)
	// }
	logger := zap.NewExample(zap.WrapCore((&apmzap.Core{}).WrapCore))
	return logger
}
