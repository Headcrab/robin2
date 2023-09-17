// todo: configurable log levels (console, file)
// todo: log size limit
// todo: log rotation time
// todo: log name by start time
// todo: change to slog
package logger

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// log singleton
var (
	//  lock = &sync.Mutex{}
	//  fileLock = &sync.Mutex{}
	once          sync.Once
	onceFile      sync.Once
	consoleLogPtr *zerolog.Logger
	fileLogPtr    *zerolog.Logger
	file          *os.File
)

type LogLevel int

// const (
// 	Trace = iota
// 	Debug
// 	Info
// 	Warn
// 	Error
// 	Fatal
// )

func consoleLog() *zerolog.Logger {
	// lock.Lock()
	// defer lock.Unlock()
	once.Do(func() {

		tmp := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "02.01.2006 15:04:05"}).
			// Level(zerolog.TraceLevel).
			Level(zerolog.DebugLevel).
			// Level(zerolog.InfoLevel).
			With().Timestamp().Logger()
		consoleLogPtr = &tmp
		consoleLogPtr.Debug().Msg("Initializing logger")
	})
	return consoleLogPtr
}

func fileLog() *zerolog.Logger {
	// fileLock.Lock()
	// defer fileLock.Unlock()
	onceFile.Do(func() {
		checkFileLog()
		// defer file.Close()
		tmp := zerolog.New(zerolog.ConsoleWriter{Out: file, TimeFormat: "02.01.2006 15:04:05", NoColor: true}).
			Level(zerolog.TraceLevel).
			// Level(zerolog.DebugLevel).
			// Level(zerolog.InfoLevel).
			With().Timestamp().Logger()
		fileLogPtr = &tmp
		// fileLogPtr.Debug().Msg("Initializing file logger")
	})
	return fileLogPtr
}

func isPathExists(s string) bool {
	_, err := os.Stat(s)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func Trace(msg string) {
	consoleLog().Trace().Msg(msg)
	fileLog().Trace().Msg(msg)
}

func Debug(msg string) {
	consoleLog().Debug().Msg(msg)
	fileLog().Debug().Msg(msg)
}

func Info(msg string) {
	consoleLog().Info().Msg(msg)
	fileLog().Info().Msg(msg)
}

func Warn(msg string) {
	consoleLog().Warn().Msg(msg)
	fileLog().Warn().Msg(msg)
}

func Error(msg string) {
	consoleLog().Error().Msg(msg)
	fileLog().Error().Msg(msg)
}

func Fatal(msg string) {
	consoleLog().Fatal().Msg(msg)
	fileLog().Fatal().Msg(msg)
}

func checkFileLog() {
	var logPath = "./"
	var logName = "robin" + time.Now().Format("_2006_01_02_15_04_05") + ".log"
	// cp, _ := os.Getwd()
	// fmt.Println("currPath: " + cp)
	if file == nil {
		var logPathes = []string{"../bin/log/", "../../bin/log/", "./log/", "../log/", "../../log/"}
		for _, path := range logPathes {
			if isPathExists(path) {
				logPath = path
				var err error
				file, err = os.OpenFile(logPath+logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					consoleLog().Fatal().Err(err).Msg("Error opening log file")
				}
			}
		}
	} else {
		logName = file.Name()
		// file.Close()
		s, _ := os.Stat(logName)
		if s.Size() > 1000000 {
			// file.Close()
			logName = "robin" + time.Now().Format("_2006_01_02_15_04_05") + ".log"
		}
		var err error
		file, err = os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			consoleLog().Fatal().Err(err).Msg("Error opening log file")
		}
	}
}

// opens 'file *os.File' and returns it
func GetLogHistory() ([]string, error) {
	checkFileLog()
	defer file.Close()
	f, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}
	// split to []string by new line
	s := strings.Split(strings.TrimSuffix(string(f), "\n"), "\n")
	return s, nil
}
