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

package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	adminapi "github.com/fatedier/frp/client/http"
	"github.com/fatedier/frp/client/proxy"
	httppkg "github.com/fatedier/frp/pkg/util/http"
	"github.com/fatedier/frp/pkg/util/log"
	netpkg "github.com/fatedier/frp/pkg/util/net"
	"github.com/fatedier/frp/pkg/util/util"
)

// clientCaptchaStore stores captcha codes temporarily for client admin API.
var clientCaptchaStore = struct {
	sync.Mutex
	m map[string]string
}{m: make(map[string]string)}

func (svr *Service) registerRouteHandlers(helper *httppkg.RouterRegisterHelper) {
	apiController := newAPIController(svr)

	// Public endpoints (no auth required)
	helper.Router.HandleFunc("/healthz", healthz)
	helper.Router.HandleFunc("/api/login", svr.apiLogin).Methods("POST")
	helper.Router.HandleFunc("/api/logout", svr.apiLogout).Methods("POST")
	helper.Router.HandleFunc("/api/captcha", svr.apiCaptcha).Methods("GET")
	helper.Router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})

	// API routes and static files with auth
	subRouter := helper.Router.NewRoute().Subrouter()
	subRouter.Use(helper.AuthMiddleware)
	subRouter.Use(httppkg.NewRequestLogger)
	subRouter.HandleFunc("/api/reload", httppkg.MakeHTTPHandlerFunc(apiController.Reload)).Methods(http.MethodGet)
	subRouter.HandleFunc("/api/stop", httppkg.MakeHTTPHandlerFunc(apiController.Stop)).Methods(http.MethodPost)
	subRouter.HandleFunc("/api/status", httppkg.MakeHTTPHandlerFunc(apiController.Status)).Methods(http.MethodGet)
	subRouter.HandleFunc("/api/config", httppkg.MakeHTTPHandlerFunc(apiController.GetConfig)).Methods(http.MethodGet)
	subRouter.HandleFunc("/api/config", httppkg.MakeHTTPHandlerFunc(apiController.PutConfig)).Methods(http.MethodPut)
	subRouter.HandleFunc("/api/proxy/{name}/config", httppkg.MakeHTTPHandlerFunc(apiController.GetProxyConfig)).Methods(http.MethodGet)
	subRouter.HandleFunc("/api/visitor/{name}/config", httppkg.MakeHTTPHandlerFunc(apiController.GetVisitorConfig)).Methods(http.MethodGet)

	if svr.storeSource != nil {
		subRouter.HandleFunc("/api/store/proxies", httppkg.MakeHTTPHandlerFunc(apiController.ListStoreProxies)).Methods(http.MethodGet)
		subRouter.HandleFunc("/api/store/proxies", httppkg.MakeHTTPHandlerFunc(apiController.CreateStoreProxy)).Methods(http.MethodPost)
		subRouter.HandleFunc("/api/store/proxies/{name}", httppkg.MakeHTTPHandlerFunc(apiController.GetStoreProxy)).Methods(http.MethodGet)
		subRouter.HandleFunc("/api/store/proxies/{name}", httppkg.MakeHTTPHandlerFunc(apiController.UpdateStoreProxy)).Methods(http.MethodPut)
		subRouter.HandleFunc("/api/store/proxies/{name}", httppkg.MakeHTTPHandlerFunc(apiController.DeleteStoreProxy)).Methods(http.MethodDelete)
		subRouter.HandleFunc("/api/store/visitors", httppkg.MakeHTTPHandlerFunc(apiController.ListStoreVisitors)).Methods(http.MethodGet)
		subRouter.HandleFunc("/api/store/visitors", httppkg.MakeHTTPHandlerFunc(apiController.CreateStoreVisitor)).Methods(http.MethodPost)
		subRouter.HandleFunc("/api/store/visitors/{name}", httppkg.MakeHTTPHandlerFunc(apiController.GetStoreVisitor)).Methods(http.MethodGet)
		subRouter.HandleFunc("/api/store/visitors/{name}", httppkg.MakeHTTPHandlerFunc(apiController.UpdateStoreVisitor)).Methods(http.MethodPut)
		subRouter.HandleFunc("/api/store/visitors/{name}", httppkg.MakeHTTPHandlerFunc(apiController.DeleteStoreVisitor)).Methods(http.MethodDelete)
	}

	subRouter.Handle("/favicon.ico", http.FileServer(helper.AssetsFS)).Methods("GET")
	subRouter.PathPrefix("/static/").Handler(
		netpkg.MakeHTTPGzipHandler(http.StripPrefix("/static/", http.FileServer(helper.AssetsFS))),
	).Methods("GET")
	subRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusMovedPermanently)
	})
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func newAPIController(svr *Service) *adminapi.Controller {
	manager := newServiceConfigManager(svr)
	return adminapi.NewController(adminapi.ControllerParams{
		ServerAddr: svr.common.ServerAddr,
		Manager:    manager,
	})
}

// getAllProxyStatus returns all proxy statuses.
func (svr *Service) getAllProxyStatus() []*proxy.WorkingStatus {
	svr.ctlMu.RLock()
	ctl := svr.ctl
	svr.ctlMu.RUnlock()
	if ctl == nil {
		return nil
	}
	return ctl.pm.GetAllProxyStatus()
}

type clientLoginReq struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	CaptchaID  string `json:"captchaId"`
	CaptchaAns string `json:"captchaAns"`
}

func clientSha256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (svr *Service) apiLogin(w http.ResponseWriter, r *http.Request) {
	var req clientLoginReq
	_ = json.NewDecoder(r.Body).Decode(&req)

	// Verify captcha
	if req.CaptchaID != "" {
		clientCaptchaStore.Lock()
		ans, ok := clientCaptchaStore.m[req.CaptchaID]
		if ok {
			delete(clientCaptchaStore.m, req.CaptchaID)
		}
		clientCaptchaStore.Unlock()
		if !ok || ans != req.CaptchaAns {
			log.Infof("captcha verification failed: captchaId=%s", req.CaptchaID)
			http.Error(w, "invalid captcha", http.StatusUnauthorized)
			return
		}
	}

	// Reject if no credentials configured
	if svr.common.WebServer.User == "" && svr.common.WebServer.Password == "" {
		http.Error(w, "authentication not configured", http.StatusUnauthorized)
		return
	}

	// Verify username and password (password is SHA256 hashed from client)
	expectedPasswordHash := clientSha256Hash(svr.common.WebServer.Password)
	if req.Username != svr.common.WebServer.User || req.Password != expectedPasswordHash {
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
	code := fmt.Sprintf("%04d", rand.Intn(10000))
	clientCaptchaStore.Lock()
	clientCaptchaStore.m[id] = code
	clientCaptchaStore.Unlock()

	svg := "<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"80\" height=\"40\">" +
		"<rect width=\"80\" height=\"40\" fill=\"#f2f2f2\"/>" +
		"<text x=\"18\" y=\"27\" font-size=\"20\" fill=\"#333\">" + code + "</text></svg>"
	resp := map[string]string{"id": id, "svg": svg}
	buf, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(buf)
}
