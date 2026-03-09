package robin

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	templateNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]+$`)
	templateLikePattern = regexp.MustCompile(`^[A-Za-z0-9_.:%-]*$`)
	templateArgKey      = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

func (a *App) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if a.adminToken == "" {
		http.Error(w, "Admin token is not configured", http.StatusServiceUnavailable)
		return false
	}

	token := strings.TrimSpace(r.Header.Get("X-Admin-Token"))
	if token == "" {
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			token = strings.TrimSpace(authHeader[7:])
		}
	}

	if subtle.ConstantTimeCompare([]byte(token), []byte(a.adminToken)) != 1 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}

	return true
}

func (a *App) requireAdminOrSameOrigin(w http.ResponseWriter, r *http.Request) bool {
	if a.hasValidAdminToken(r) || isSameOriginRequest(r) {
		return true
	}

	if a.adminToken == "" {
		http.Error(w, "Admin token is not configured", http.StatusServiceUnavailable)
		return false
	}

	http.Error(w, "Forbidden", http.StatusForbidden)
	return false
}

func (a *App) hasValidAdminToken(r *http.Request) bool {
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

func isSameOriginRequest(r *http.Request) bool {
	return sameOriginHost(r.Header.Get("Origin"), r.Host) || sameOriginHost(r.Header.Get("Referer"), r.Host)
}

func sameOriginHost(rawURL, host string) bool {
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

func requireMethod(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	for _, method := range methods {
		if r.Method == method {
			return true
		}
	}

	w.Header().Set("Allow", strings.Join(methods, ", "))
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return false
}

func isSafeTemplateName(name string) bool {
	return templateNamePattern.MatchString(name)
}

func isSafeTemplateLike(like string) bool {
	return templateLikePattern.MatchString(like)
}

func isSafeTemplateArgKey(key string) bool {
	return templateArgKey.MatchString(key)
}
