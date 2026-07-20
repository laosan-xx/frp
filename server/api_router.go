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
	"sync"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	httppkg "github.com/fatedier/frp/pkg/util/http"
	"github.com/fatedier/frp/pkg/util/log"
	netpkg "github.com/fatedier/frp/pkg/util/net"
	"github.com/fatedier/frp/pkg/util/util"
	adminapi "github.com/fatedier/frp/server/http"
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

	apiController := adminapi.NewController(svr.cfg, svr.clientRegistry, svr.pxyManager)

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
	subRouter.HandleFunc("/api/v2/proxies", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ProxyList)).Methods("GET")
	v2EncodedPathRouter.HandleFunc("/api/v2/proxies/{name}/traffic", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ProxyTraffic)).Methods("GET")
	v2EncodedPathRouter.HandleFunc("/api/v2/proxies/{name}", httppkg.MakeHTTPHandlerFuncV2(apiController.APIV2ProxyDetail)).Methods("GET")

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

func (svr *Service) apiCaptcha(w http.ResponseWriter, _ *http.Request) {
	id, _ := util.RandID()
	// Generate 4-digit captcha code
	code := fmt.Sprintf("%04d", rand.Intn(10000))
	captchaStore.Lock()
	captchaStore.m[id] = code
	captchaStore.Unlock()

	svg := "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"80\" height=\"40\">" +
		"<rect width=\"80\" height=\"40\" fill=\"#f2f2f2\"/>" +
		"<text x=\"18\" y=\"27\" font-size=\"20\" fill=\"#333\">" + code + "</text></svg>"
	resp := map[string]string{"id": id, "svg": svg}
	buf, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(buf)
}
