package logger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/joho/godotenv"
)

var (
	once          sync.Once
	onceFile      sync.Once
	consoleLogPtr *slog.Logger
	fileLogPtr    *slog.Logger
	file          *os.File
)

type LogLevel int

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// переменные, объявленные для извлечения значений из .env
var (
	logRotationSize    int64
	logRotationTime    time.Duration
	logPath            string
	defaultLogFileName string
	logLevel           string
)

// инициализация переменных из .env файла
func init() {
	err := godotenv.Load(filepath.Join(GetWorkDir(), ".env"))
	if err != nil {
		slog.Error("Error loading .env file")
	}

	logRotationSize, err = strconv.ParseInt(getEnv("LOG_ROTATION_SIZE", "100000"), 10, 64)
	if err != nil {
		slog.Error("Error parsing LOG_ROTATION_SIZE", "error", err.Error())
	}

	logRotationTime, err = time.ParseDuration(getEnv("LOG_ROTATION_TIME", "24h"))
	if err != nil {
		slog.Error("Error parsing LOG_ROTATION_TIME", "error", err.Error())
	}

	logPath = getEnv("LOG_PATH", "./log/")
	defaultLogFileName = getEnv("LOG_DEFAULT_FILE_NAME", "robin.log")
	logLevel = getEnv("LOG_LEVEL", "info")

}

// вспомогательная функция для получения значения переменной из .env файла
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetWorkDir() string {
	executablePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		slog.Debug(err.Error())
		return ""
	}

	dir := filepath.Dir(executablePath)
	slog.Debug("working dir set to: " + dir)

	return dir
}

func getLogLevel(level LogLevel) slog.Level {
	switch level {
	case LevelTrace:
		return slog.LevelDebug - 1
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelFatal:
		return slog.LevelError + 1
	default:
		return slog.LevelInfo
	}
}

func consoleLog() *slog.Logger {
	once.Do(func() {
		consoleLogPtr = slog.New(
			slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: getLogLevel(strToLogLevel(logLevel))}))
		consoleLogPtr.Info("Initializing console logger")
	})
	return consoleLogPtr
}

func fileLog() *slog.Logger {
	onceFile.Do(func() {
		checkFileLog()
		fileLogPtr = slog.New(
			slog.NewJSONHandler(file, &slog.HandlerOptions{Level: getLogLevel(strToLogLevel(logLevel))}))
		fileLogPtr.Info("Initializing file logger")
	})
	return fileLogPtr
}

func isPathExists(s string) bool {
	_, err := os.Stat(s)
	return err == nil || os.IsExist(err)
}

// методы логирования для различного уровня логов
func Trace(msg string) {
	logMessage(LevelTrace, msg)
}

func Debug(msg string) {
	logMessage(LevelDebug, msg)
}

func Info(msg string) {
	logMessage(LevelInfo, msg)
}

func Warn(msg string) {
	logMessage(LevelWarn, msg)
}

func Error(msg string) {
	logMessage(LevelError, msg)
}

func Fatal(msg string) {
	logMessage(LevelFatal, msg)
}

func logMessage(level LogLevel, msg string) {
	consoleLog().LogAttrs(context.Background(), getLogLevel(level), msg)
	fileLog().LogAttrs(context.Background(), getLogLevel(level), msg)
}

func checkFileLog() {
	closed := checkFileClosed(file)
	if file == nil || closed || isFileSizeExceeded(file) || isTimeExceeded(file) {
		file = createNewLogFile()
	}
}

func checkFileClosed(f *os.File) bool {
	if f == nil {
		return true
	}
	_, err := f.Stat()
	return err != nil
}

func isFileSizeExceeded(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		consoleLog().Error("Cannot get file info", "error", err.Error())
		return false
	}
	return info.Size() > logRotationSize
}

func isTimeExceeded(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		consoleLog().Error("Cannot get file info", "error", err.Error())
		return false
	}
	modTime := info.ModTime()
	return time.Since(modTime) > logRotationTime
}

func createNewLogFile() *os.File {
	if !isPathExists(logPath) {
		err := os.MkdirAll(logPath, os.ModePerm)
		if err != nil {
			consoleLog().Error("Failed to create log directory", "error", err.Error())
		}
	}
	logFileName := logPath + getLogFileName()
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		consoleLog().Error("Error opening log file", "error", err.Error())
	}
	return f
}

func getLogFileName() string {
	return defaultLogFileName
}

type LogItem struct {
	Date  time.Time
	Msg   string
	Level string
}

type LogHistory []LogItem

func GetLogHistory() (LogHistory, error) {
	checkFileLog()
	if file == nil {
		return LogHistory{}, nil
	}

	// читаем файл по имени, не закрывая дескриптор
	logFileName := logPath + getLogFileName()
	f, err := os.ReadFile(logFileName)
	if err != nil {
		return LogHistory{}, err
	}
	return parseLog(string(f)), nil
}

func parseLog(log string) []LogItem {
	var logItems []LogItem
	lines := strings.Split(log, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		logItems = append(logItems, parseLogLine(line))
	}
	return logItems
}

func parseLogLine(line string) LogItem {
	// parse json string
	var logItem LogItem
	err := json.Unmarshal([]byte(line), &logItem)
	if err == nil {
		return logItem
	}

	// если JSON парсинг не удался, попробуем парсить строчный формат
	// предполагаемый формат: time=2006-01-02T15:04:05.000+05:00 level=INFO msg="message"
	parts := strings.Fields(line)
	logItem = LogItem{
		Date:  time.Now(), // используем текущее время как fallback
		Level: "UNKNOWN",
		Msg:   line, // весь текст строки как сообщение
	}

	// ищем msg= и парсим все что после него как сообщение
	msgStartIndex := strings.Index(line, "msg=")
	if msgStartIndex != -1 {
		msgStr := line[msgStartIndex+4:] // +4 для "msg="
		// убираем кавычки если есть
		if len(msgStr) >= 2 && msgStr[0] == '"' && msgStr[len(msgStr)-1] == '"' {
			msgStr = msgStr[1 : len(msgStr)-1]
		}
		logItem.Msg = msgStr
	}

	for _, part := range parts {
		if strings.HasPrefix(part, "time=") {
			timeStr := strings.TrimPrefix(part, "time=")
			if parsedTime, err := time.Parse(time.RFC3339, timeStr); err == nil {
				logItem.Date = parsedTime
			}
		} else if strings.HasPrefix(part, "level=") {
			logItem.Level = strings.TrimPrefix(part, "level=")
		}
	}

	return logItem
}

func strToLogLevel(level string) LogLevel {
	switch level {
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// ClearLogs очищает файл логов
func ClearLogs() error {
	// получаем путь к файлу логов
	logFileName := logPath + getLogFileName()

	// закрываем текущий файл если он открыт
	if file != nil {
		file.Close()
		file = nil
	}

	// сбрасываем логгер для переинициализации
	fileLogPtr = nil
	onceFile = sync.Once{}

	// очищаем содержимое файла (truncate) вместо удаления
	f, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		consoleLog().Error("Failed to truncate log file", "error", err.Error())
		return err
	}
	f.Close()

	// инициализируем новый файл логов
	checkFileLog()

	Info("Логи очищены")
	return nil
}
