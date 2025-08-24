package logger

import (
	"os"

	"github.com/williamkoller/golang-payment-stripe/internal/infra/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg *config.Config) *zap.Logger {
	level := zapcore.InfoLevel
	_ = level.UnmarshalText([]byte(cfg.LogLevel))
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), zapcore.AddSync(os.Stdout), level)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}
