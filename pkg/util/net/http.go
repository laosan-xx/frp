// Copyright 2017 fatedier, fatedier@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net

import (
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/laosan-xx/frp/pkg/util/util"
)

// isStaticFileRequest 检查是否为静态文件请求
func isStaticFileRequest(path string) bool {
	// 允许的静态文件路径
	staticPaths := []string{
		"/static/",
		"/favicon.ico",
	}

	for _, staticPath := range staticPaths {
		if path == staticPath || strings.HasPrefix(path, staticPath) {
			return true
		}
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(path))
	staticExts := []string{".html", ".css", ".js", ".ico", ".png", ".jpg", ".gif", ".svg", ".woff", ".woff2", ".ttf", ".eot"}
	for _, staticExt := range staticExts {
		if ext == staticExt {
			return true
		}
	}

	return false
}

type HTTPAuthMiddleware struct {
	user          string
	passwd        string
	authFailDelay time.Duration
}

func NewHTTPAuthMiddleware(user, passwd string) *HTTPAuthMiddleware {
	return &HTTPAuthMiddleware{
		user:   user,
		passwd: passwd,
	}
}

func (authMid *HTTPAuthMiddleware) SetAuthFailDelay(delay time.Duration) *HTTPAuthMiddleware {
	authMid.authFailDelay = delay
	return authMid
}

func (authMid *HTTPAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查是否为静态文件路径，如果是则直接放行
		if isStaticFileRequest(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// 安全检查：如果未设置用户名和密码，拒绝访问
		if authMid.user == "" && authMid.passwd == "" {
			if authMid.authFailDelay > 0 {
				time.Sleep(authMid.authFailDelay)
			}
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "authentication not configured", http.StatusUnauthorized)
			return
		}

		reqUser, reqPasswd, hasAuth := r.BasicAuth()
		if hasAuth && util.ConstantTimeEqString(reqUser, authMid.user) &&
			util.ConstantTimeEqString(reqPasswd, authMid.passwd) {
			next.ServeHTTP(w, r)
		} else {
			if authMid.authFailDelay > 0 {
				time.Sleep(authMid.authFailDelay)
			}
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
	})
}

// SessionManager provides cookie-based auth for dashboard.
type SessionManager struct {
	secret     []byte
	cookieName string
	ttl        time.Duration
	sameSite   http.SameSite
	secure     bool
}

type sessionPayload struct {
	U string `json:"u"`
	E int64  `json:"e"`
	N string `json:"n"`
}

func NewSessionManager(secret []byte, cookieName string, ttl time.Duration, sameSite http.SameSite, secure bool) *SessionManager {
	return &SessionManager{secret: secret, cookieName: cookieName, ttl: ttl, sameSite: sameSite, secure: secure}
}

func (sm *SessionManager) sign(b []byte) string {
	mac := hmac.New(sha256.New, sm.secret)
	mac.Write(b)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (sm *SessionManager) encode(p sessionPayload) (string, error) {
	jb, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	sig := sm.sign(jb)
	token := base64.RawURLEncoding.EncodeToString(jb) + "." + sig
	return token, nil
}

func (sm *SessionManager) decode(token string) (sessionPayload, bool) {
	var out sessionPayload
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return out, false
	}
	jb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return out, false
	}
	if !hmac.Equal([]byte(sm.sign(jb)), []byte(parts[1])) {
		return out, false
	}
	if err = json.Unmarshal(jb, &out); err != nil {
		return out, false
	}
	if out.E < time.Now().Unix() {
		return out, false
	}
	return out, true
}

func (sm *SessionManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查是否为静态文件路径，如果是则直接放行
		if isStaticFileRequest(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		c, err := r.Cookie(sm.cookieName)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if _, ok := sm.decode(c.Value); !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (sm *SessionManager) Issue(w http.ResponseWriter, username string) error {
	p := sessionPayload{U: username, E: time.Now().Add(sm.ttl).Unix(), N: strconv.FormatInt(time.Now().UnixNano(), 10)}
	token, err := sm.encode(p)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     sm.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   sm.secure,
		SameSite: sm.sameSite,
		Expires:  time.Now().Add(sm.ttl),
		MaxAge:   int(sm.ttl.Seconds()),
	}
	http.SetCookie(w, cookie)
	return nil
}

func (sm *SessionManager) Clear(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sm.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   sm.secure,
		SameSite: sm.sameSite,
	}
	http.SetCookie(w, cookie)
}

type HTTPGzipWrapper struct {
	h http.Handler
}

func (gw *HTTPGzipWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		gw.h.ServeHTTP(w, r)
		return
	}
	w.Header().Set("Content-Encoding", "gzip")
	gz := gzip.NewWriter(w)
	defer gz.Close()
	gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
	gw.h.ServeHTTP(gzr, r)
}

func MakeHTTPGzipHandler(h http.Handler) http.Handler {
	return &HTTPGzipWrapper{
		h: h,
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
