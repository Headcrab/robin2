package utils

import (
	"net"
	"os"
	"path/filepath"
	"robin2/internal/errors"
	"robin2/internal/logger"
	"strconv"
	"strings"
	"time"
)

func GetWorkDir() string {
	executablePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logger.Trace(err.Error())
		return ""
	}

	dir := filepath.Dir(executablePath)
	logger.Trace("working dir set to: " + dir)

	return dir
}

func GetLocalhostIpAdresses() []string {
	localhostIPs := []string{"127.0.0.1"}
	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Error(err.Error())
		return localhostIPs
	}
	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					localhostIPs = append(localhostIPs, ip4.String())
				}
			}
		}
	}
	return localhostIPs
}

// thenIf - функция, которая принимает условие, значение ifTrue и значение ifFalse в качестве параметров.
// Она возвращает значение ifTrue, если условие истинно, и значение ifFalse в противном случае.
// Функция может работать с любым типом входных данных.
func ThenIf[T any](condition bool, ifTrue T, ifFalse T) T {
	if condition {
		return ifTrue
	}
	return ifFalse
}

// excelTimeToTime преобразует строку времени из Excel в объект времени time.Time в Go.
//
// В качестве параметра принимается строка timeStr, представляющая значение времени в Excel.
// Функция возвращает объект time.Time и ошибку.
// excelTimeToTime converts a time string from Excel to a time.Time object in Go.
//
// It takes a timeStr string as a parameter, representing the time value in Excel.
// The function returns a time.Time object and an error.
func ExcelTimeToTime(timeStr string, formats []string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, errors.ErrInvalidDate
	}

	parsed, err := parseTime(timeStr, formats)
	if err != nil {
		return time.Time{}, err
	}

	return parsed, nil
}

func parseTime(timeStr string, formats []string) (time.Time, error) {
	if !strings.Contains(timeStr, ":") {
		timeStr = strings.Replace(timeStr, ",", ".", 1)
		timeFloat, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			return time.Time{}, errors.ErrNotAFloat
		}

		unixTime := (timeFloat - 25569.0) * 86400.0
		return time.Unix(int64(unixTime), 0.0).UTC(), nil
	}

	tm, err := TryParseDate(timeStr, formats)
	if err != nil {
		return time.Time{}, err
	}
	return tm.Local(), nil
}

// tryParseDate пытается разобрать строку в качестве даты и возвращает разобранную дату в случае успеха.
// Если входная строка не может быть разобрана как дата, функция возвращает ошибку.
//
// Параметры:
// - input: Строковый вход, который нужно разобрать как дату.
//
// Возвращает:
// - time.Time: Разобранная дата
func TryParseDate(date string, formats []string) (time.Time, error) {
	// if date is empty, return error
	if date == "" {
		return time.Time{}, errors.ErrInvalidDate
	}
	// if date is not empty, try to parse it to time.Time
	// if date is not valid, return error
	// cfg := a.config.GetStringSlice("app.date_formats")
	for fm := range formats {
		t, err := time.ParseInLocation(formats[fm], date, time.Local)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.ErrInvalidDate
}
