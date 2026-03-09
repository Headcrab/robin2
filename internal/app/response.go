package robin

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"robin2/internal/logger"
	"strings"
	"syscall"
)

func writeResponse(w http.ResponseWriter, body []byte) bool {
	_, err := w.Write(body)
	if err != nil {
		logResponseWriteError(err)
		return false
	}
	return true
}

func writeStringResponse(w http.ResponseWriter, body string) bool {
	return writeResponse(w, []byte(body))
}

func logResponseWriteError(err error) {
	if err == nil {
		return
	}
	if isClientDisconnect(err) {
		logger.Debug(fmt.Sprintf("client closed connection while writing response: %v", err))
		return
	}
	logger.Error(fmt.Sprintf("Error writing response: %v", err))
}

func isClientDisconnect(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, net.ErrClosed) {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) && isClientDisconnect(opErr.Err) {
		return true
	}

	var sysErr *os.SyscallError
	if errors.As(err, &sysErr) && isClientDisconnect(sysErr.Err) {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "broken pipe") || strings.Contains(msg, "connection reset by peer")
}
