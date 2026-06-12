// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0

package faces

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	sessionCookieName = "faces-admin-session"
	sessionShortTTL   = 8 * time.Hour
	sessionLongTTL    = 7 * 24 * time.Hour
)

// sessionStore is an in-memory token store. Tokens are random 32-char hex strings.
// The store is safe for concurrent use and self-cleans expired entries every 10 minutes.
type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]time.Time
}

func newSessionStore() *sessionStore {
	s := &sessionStore{sessions: make(map[string]time.Time)}
	go s.gcLoop()
	return s
}

func (s *sessionStore) create(ttl time.Duration) string {
	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)
	s.mu.Lock()
	s.sessions[token] = time.Now().Add(ttl)
	s.mu.Unlock()
	return token
}

func (s *sessionStore) valid(token string) bool {
	s.mu.RLock()
	exp, ok := s.sessions[token]
	s.mu.RUnlock()
	return ok && time.Now().Before(exp)
}

func (s *sessionStore) delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *sessionStore) gcLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		for t, exp := range s.sessions {
			if time.Now().After(exp) {
				delete(s.sessions, t)
			}
		}
		s.mu.Unlock()
	}
}

// authMiddleware wraps next with session-cookie authentication.
// If username is empty, auth is disabled and next is returned unchanged.
func authMiddleware(next http.Handler, sessions *sessionStore, username, password string) http.Handler {
	if username == "" || password == "" {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Always open: health probe, login page, login/logout API
		if path == "/healthz" || path == "/login" || path == "/api/login" || path == "/api/logout" {
			next.ServeHTTP(w, r)
			return
		}
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !sessions.valid(cookie.Value) {
			if strings.HasPrefix(path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			} else {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handleLogin serves GET /login (login page) and POST /api/login (credential check).
func (a *AdminProvider) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		a.serveLoginPage(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		RememberMe bool   `json:"rememberMe"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if body.Username != a.adminUsername || body.Password != a.adminPassword {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
		return
	}

	ttl := sessionShortTTL
	if body.RememberMe {
		ttl = sessionLongTTL
	}
	token := a.sessions.create(ttl)

	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if body.RememberMe {
		cookie.MaxAge = int(ttl.Seconds())
	}
	http.SetCookie(w, cookie)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleLogout clears the session and redirects to the login page.
func (a *AdminProvider) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		a.sessions.delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// serveLoginPage reads login.html from DATA_PATH and serves it.
func (a *AdminProvider) serveLoginPage(w http.ResponseWriter, r *http.Request) {
	candidate := filepath.Join(a.dataPath, "login.html")
	raw, err := os.ReadFile(candidate)
	if err != nil {
		http.Error(w, "login page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
