package logger

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"
)

type Logger struct {
	logger *slog.Logger
	mu     sync.Mutex
}

func NewFileLogger(filename string, level slog.Level) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	handler := NewMultilineHandler(file, &slog.HandlerOptions{Level: level})
	return &Logger{logger: slog.New(handler)}, nil
}

func (l *Logger) initIfNeeded() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logger == nil {
		l.logger = CreateDefaultLogger().logger
	}
}

func GetDefaultLogger() *Logger {
	return &Logger{}
}

func IsDebug() bool {
	return os.Getenv("GRAPH_LOG_LEVEL") == "debug"
}

func getLogLevel() slog.Level {
	if IsDebug() {
		return slog.LevelDebug
	}
	return slog.LevelWarn
}

func CreateDefaultLogger() *Logger {
	logLevel := getLogLevel()

	logDir := fmt.Sprintf("%s/.git-graph/log", os.Getenv("HOME"))
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		os.Mkdir(logDir, 0755)
	}

	filename := fmt.Sprintf("%s/%s.log", logDir, time.Now().Format("2006-01-02-15-04-05"))
	logger, err := NewFileLogger(filename, logLevel.Level())
	if err != nil {
		panic(err)
	}
	return logger
}

func (l *Logger) Fatal(msg string) {
	if getLogLevel() > slog.LevelError {
		return
	}
	l.initIfNeeded()
	l.logger.Error(msg)
	log.Fatal(msg)
}

func (l *Logger) Debug(msg string) {
	if getLogLevel() > slog.LevelDebug {
		return
	}
	l.initIfNeeded()
	l.logger.Debug(msg)
}
