package logger

import "go.uber.org/zap"

// Log - global logger instance
var Log *zap.Logger

// Init - configure the logger
// development=true: pretty, colorized log output (for terminal)
// development=false: JSON log output (for production, convenient for Loki/ELK)
func Init(development bool) {
	var err error
	if development {
		Log, err = zap.NewDevelopment()
	} else {
		Log, err = zap.NewProduction()
	}
	if err != nil {
		panic("logger yaratib bo'lmadi: " + err.Error())
	}
}

// Sync - flush any buffered log entries (used with defer)
func Sync() {
	if Log != nil {
		Log.Sync()
	}
}
