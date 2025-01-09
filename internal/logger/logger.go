package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// BuildLogger строит zap.Logger с необходимым уровнем логирования.
func BuildLogger(level string) (*zap.Logger, error) {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse atomic level: %w", err)
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()

	// Настройка формата времени
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	// Отключение stacktrace
	cfg.DisableStacktrace = true

	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build config %w", err)
	}
	return zl, nil
}
