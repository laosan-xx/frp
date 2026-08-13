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

package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"

	httppkg "github.com/laosan-xx/frp/pkg/util/http"
	"github.com/laosan-xx/frp/pkg/util/log"
	netpkg "github.com/laosan-xx/frp/pkg/util/net"
	"github.com/laosan-xx/frp/pkg/util/util"
	adminapi "github.com/laosan-xx/frp/server/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// captchaStore stores captcha codes temporarily.
var captchaStore = struct {
	sync.Mutex
	m map[string]string
}{m: make(map[string]string)}

func (svr *Service) registerRouteHandlers(helper *httppkg.RouterRegisterHelper) {
	// Public endpoints (no auth required)
	helper.Router.HandleFunc("/healthz", healthz)
	helper.Router.HandleFunc("/api/login", svr.apiLogin).Methods("POST")
	helper.Router.HandleFunc("/api/logout", svr.apiLogout).Methods("POST")
	helper.Router.HandleFunc("/api/captcha", svr.apiCaptcha).Methods("GET")
	helper.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})

	subRouter := helper.Router.NewRoute().Subrouter()

	subRouter.Use(helper.AuthMiddleware)
	subRouter.Use(httppkg.NewRequestLogger)

	// metrics
	if svr.cfg.EnablePrometheus {
		subRouter.Handle("/metrics", promhttp.Handler())
	}

	apiController := adminapi.NewController(svr.cfg, svr.clientRegistry, svr.pxyManager, svr.ctlManager, svr.ipLookup)

	// apis
	subRouter.HandleFunc("/api/serverinfo", httppkg.MakeHTTPHandlerFunc(apiController.APIServerInfo)).Methods("GET")
	subRouter.HandleFunc("/api/proxy/{type}", httppkg.MakeHTTPHandlerFunc(apiController.APIProxyByType)).Methods("GET")
	subRouter.HandleFunc("/api/proxy/{type}/{name}", httppkg.MakeHTTPHandlerFunc(apiController.APIProxyByTypeAndName)).Methods("GET")
	subRouter.HandleFunc("/api/proxies/{name}", httppkg.MakeHTTPHandlerFunc(apiController.APIProxyByName)).Methods("GET")
	subRouter.HandleFunc("/api/traffic/{name}", httppkg.MakeHTTPHandlerFunc(apiController.APIProxyTraffic)).Methods("GET")
	subRouter.HandleFunc("/api/clients", httppkg.MakeHTTPHandlerFunc(apiController.APIClientList)).Methods("GET")
	subRouter.HandleFunc("/api/clients/{key}", httppkg.MakeHTTPHandlerFunc(apiController.APIClientDetail)).Methods("GET")
	subRouter.HandleFunc("/api/proxies", httppkg.MakeHTTPHandlerFunc(apiController.DeleteProxies)).Methods("DELETE")

	subRouter.HandleFunc("/api/v2/users", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2UserList)).Methods("GET")
	subRouter.HandleFunc("/api/v2/system/info", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2SystemInfo)).Methods("GET")
	subRouter.HandleFunc("/api/v2/system/prune", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2SystemPrune)).Methods("POST")
	subRouter.HandleFunc("/api/v2/clients", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ClientList)).Methods("GET")
	v2EncodedPathRouter := subRouter.NewRoute().Subrouter()
	v2EncodedPathRouter.UseEncodedPath()
	v2EncodedPathRouter.HandleFunc("/api/v2/clients/{key}", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ClientDetail)).Methods("GET")
	v2EncodedPathRouter.HandleFunc("/api/v2/clients/{key}", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ClientDelete)).Methods("DELETE")
	v2EncodedPathRouter.HandleFunc("/api/v2/clients/{key}/command", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ClientCommand)).Methods("POST")

	// clientID / runID based routes (no "{user}.{clientID}" composite key needed)
	v2EncodedPathRouter.HandleFunc("/api/v2/client/{id}", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ClientDetailByID)).Methods("GET")
	v2EncodedPathRouter.HandleFunc("/api/v2/client/{id}", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ClientDeleteByID)).Methods("DELETE")
	v2EncodedPathRouter.HandleFunc("/api/v2/client/{id}/command", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ClientCommandByID)).Methods("POST")
	v2EncodedPathRouter.HandleFunc("/api/v2/client/run/{runID}", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ClientDetailByRunID)).Methods("GET")
	subRouter.HandleFunc("/api/v2/proxies", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ProxyList)).Methods("GET")
	v2EncodedPathRouter.HandleFunc("/api/v2/proxies/{id}/traffic", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ProxyTraffic)).Methods("GET")
	v2EncodedPathRouter.HandleFunc("/api/v2/proxies/{id}", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ProxyDetail)).Methods("GET")

	// Firmware API proxy (server-side GitHub API calls, bypass shared proxy rate limiting)
	subRouter.HandleFunc("/api/v2/firmware/releases", httppkg.MakeHTTPHandlerFuncV2(NewFirmwareHandler(svr.cfg.GitHubToken))).Methods("GET")

	// view
	subRouter.Handle("/favicon.ico", http.FileServer(helper.AssetsFS)).Methods("GET")
	subRouter.PathPrefix("/static/").Handler(
		netpkg.MakeHTTPGzipHandler(http.StripPrefix("/static/", http.FileServer(helper.AssetsFS))),
	).Methods("GET")

	subRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
}

type loginReq struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	CaptchaID  string `json:"captchaId"`
	CaptchaAns string `json:"captchaAns"`
}

func sha256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (svr *Service) apiLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Verify captcha
	if req.CaptchaID != "" {
		captchaStore.Lock()
		ans, ok := captchaStore.m[req.CaptchaID]
		if ok {
			delete(captchaStore.m, req.CaptchaID)
		}
		captchaStore.Unlock()
		if !ok || ans != req.CaptchaAns {
			log.Infof("captcha verification failed: captchaId=%s", req.CaptchaID)
			http.Error(w, "invalid captcha", http.StatusUnauthorized)
			return
		}
	}

	// Reject if no credentials configured
	if svr.cfg.WebServer.User == "" && svr.cfg.WebServer.Password == "" {
		http.Error(w, "authentication not configured", http.StatusUnauthorized)
		return
	}

	// Verify username and password (password is SHA256 hashed from client)
	expectedPasswordHash := sha256Hash(svr.cfg.WebServer.Password)
	if req.Username != svr.cfg.WebServer.User || req.Password != expectedPasswordHash {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	if svr.sessionMgr != nil {
		_ = svr.sessionMgr.Issue(w, req.Username)
	}
	w.WriteHeader(http.StatusOK)
}

func (svr *Service) apiLogout(w http.ResponseWriter, _ *http.Request) {
	if svr.sessionMgr != nil {
		svr.sessionMgr.Clear(w)
	}
	w.WriteHeader(http.StatusOK)
}

// captchaPalette are the character colors used for the captcha text.
var captchaPalette = []string{
	"#3b82f6", "#06b6d4", "#ef4444", "#f59e0b",
	"#10b981", "#8b5cf6", "#ec4899", "#0ea5e9",
}

// apiCaptcha generates a 4-digit captcha and returns it as a styled SVG.
func (svr *Service) apiCaptcha(w http.ResponseWriter, _ *http.Request) {
	id, _ := util.RandID()
	// Generate 4-digit captcha code
	code := fmt.Sprintf("%04d", rand.Intn(10000))
	captchaStore.Lock()
	captchaStore.m[id] = code
	captchaStore.Unlock()

	svg := buildCaptchaSVG(code)
	resp := map[string]string{"id": id, "svg": svg}
	buf, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(buf)
}

// buildCaptchaSVG renders a 4-character captcha as a polished, themed SVG
// with a soft gradient background, per-character colors, slight rotations
// and random interference lines/dots for better readability and security.
func buildCaptchaSVG(code string) string {
	const (
		w      = 86
		h      = 40
		startX = 13
		stepX  = 18
	)
	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		"<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"%d\" height=\"%d\" viewBox=\"0 0 %d %d\">",
		w, h, w, h,
	))
	// Soft gradient background
	b.WriteString("<defs><linearGradient id=\"bg\" x1=\"0\" y1=\"0\" x2=\"1\" y2=\"1\">")
	b.WriteString("<stop offset=\"0%\" stop-color=\"#eef4ff\"/>")
	b.WriteString("<stop offset=\"100%\" stop-color=\"#e0f7fa\"/>")
	b.WriteString("</linearGradient></defs>")
	b.WriteString(fmt.Sprintf("<rect width=\"%d\" height=\"%d\" rx=\"8\" fill=\"url(#bg)\"/>", w, h))

	// Interference lines
	for i := 0; i < 3; i++ {
		x1 := rand.Intn(w)
		y1 := rand.Intn(h)
		x2 := rand.Intn(w)
		y2 := rand.Intn(h)
		color := captchaPalette[rand.Intn(len(captchaPalette))]
		b.WriteString(fmt.Sprintf(
			"<line x1=\"%d\" y1=\"%d\" x2=\"%d\" y2=\"%d\" stroke=\"%s\" stroke-width=\"1\" stroke-opacity=\"0.35\"/>",
			x1, y1, x2, y2, color,
		))
	}

	// Interference dots
	for i := 0; i < 18; i++ {
		cx := rand.Intn(w)
		cy := rand.Intn(h)
		cr := rand.Intn(2) + 1
		color := captchaPalette[rand.Intn(len(captchaPalette))]
		b.WriteString(fmt.Sprintf(
			"<circle cx=\"%d\" cy=\"%d\" r=\"%d\" fill=\"%s\" fill-opacity=\"0.3\"/>",
			cx, cy, cr, color,
		))
	}

	// Characters with slight rotation and individual colors
	for i, ch := range code {
		x := startX + i*stepX
		y := 29
		rot := rand.Intn(30) - 15
		color := captchaPalette[rand.Intn(len(captchaPalette))]
		b.WriteString(fmt.Sprintf(
			"<text x=\"%d\" y=\"%d\" font-family=\"'Courier New',monospace\" font-size=\"24\" font-weight=\"700\" fill=\"%s\" transform=\"rotate(%d %d %d)\">%c</text>",
			x, y, color, rot, x, y, ch,
		))
	}
	b.WriteString("</svg>")
	return b.String()
}
