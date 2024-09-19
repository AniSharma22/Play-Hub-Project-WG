package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

var Logger *zap.SugaredLogger

func init() {
	// Create a log file
	logFile, err := os.OpenFile("api.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	// Configure the zap logger to write to both file and console
	writers := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(logFile))

	// Custom time encoder
	timeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339)) // You can use different formats here
	}

	// Configure the encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = timeEncoder

	// Create a core with custom time encoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writers,
		zapcore.InfoLevel,
	)

	zapLogger := zap.New(core)
	Logger = zapLogger.Sugar()

	// Ensure logs are flushed to file
	defer zapLogger.Sync()
}
