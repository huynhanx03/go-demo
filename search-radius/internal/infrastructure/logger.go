package infrastructure

import (
	"search-radius/global"
	"search-radius/pkg/logger"
)

// SetupLogger initializes the logger
func SetupLogger() {
	config := logger.LoggerConfig{
		Level:      global.Config.Logger.LogLevel,
		Filename:   global.Config.Logger.FileLogName,
		MaxSize:    global.Config.Logger.MaxSize,
		MaxBackups: global.Config.Logger.MaxBackups,
		MaxAge:     global.Config.Logger.MaxAge,
		Compress:   global.Config.Logger.Compress,
	}

	global.Logger = logger.NewLogger(config)
}
