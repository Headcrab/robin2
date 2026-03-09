package robin

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"strings"
)

func (a *App) requireLogClearAccess(w http.ResponseWriter, r *http.Request) bool {
	if isSameOriginLogClearRequest(r) {
		return true
	}

	if a.hasValidAdminTokenForLogClear(r) {
		return true
	}

	if a.adminToken == "" {
		http.Error(w, "Admin token is not configured", http.StatusServiceUnavailable)
		return false
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
	return false
}

func (a *App) hasValidAdminTokenForLogClear(r *http.Request) bool {
	if a.adminToken == "" {
		return false
	}

	token := strings.TrimSpace(r.Header.Get("X-Admin-Token"))
	if token == "" {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
	}

	return subtle.ConstantTimeCompare([]byte(token), []byte(a.adminToken)) == 1
}

func isSameOriginLogClearRequest(r *http.Request) bool {
	return sameOriginLogClearHost(r.Header.Get("Origin"), r.Host) || sameOriginLogClearHost(r.Header.Get("Referer"), r.Host)
}

func sameOriginLogClearHost(rawURL, host string) bool {
	rawURL = strings.TrimSpace(rawURL)
	host = strings.TrimSpace(host)
	if rawURL == "" || host == "" {
		return false
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return false
	}

	return strings.EqualFold(parsed.Host, host)
}
