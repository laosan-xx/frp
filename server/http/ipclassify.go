// Copyright 2024 fatedier, fatedier@gmail.com
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

package http

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ipClassInfo holds the classification of a public IP: geo-consistency
// (native vs broadcast), usage type (hosting/isp/business/education) and
// privacy flags. Mirrors the logic of ip_type_check.sh.
//
// This used to live in the frpc binary (client/command_builtin.go) and was
// queried from every router. It is now centralized here so that external
// IP-intel queries hit a single egress and can be cached across all clients,
// which keeps the rate-limited ip-api endpoint (45 req/min per source IP)
// from being exhausted by many routers testing many nodes.
type ipClassInfo struct {
	Country           string `json:"country"`
	RegisteredCountry string `json:"registered_country"`
	ASN               string `json:"asn"`
	Org               string `json:"org"`
	City              string `json:"city"`
	Region            string `json:"region"`
	UsageType         string `json:"usage_type"`
	CompanyType       string `json:"company_type"`
	IPType            string `json:"ip_type"`
	IsNative          bool   `json:"is_native"`
	IsBroadcast       bool   `json:"is_broadcast"`
	IsHosting         bool   `json:"is_hosting"`
	IsISP             bool   `json:"is_isp"`
	Privacy           string `json:"privacy"`
}

// ipinfoWidget is the subset of ipinfo.io/widget/demo/<IP> we consume.
type ipinfoWidget struct {
	Data struct {
		Country string `json:"country"`
		City    string `json:"city"`
		Region  string `json:"region"`
		ASN     struct {
			ASN  string `json:"asn"`
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"asn"`
		Company struct {
			Type string `json:"type"`
		} `json:"company"`
		Abuse struct {
			Country string `json:"country"`
		} `json:"abuse"`
		Privacy struct {
			Proxy   bool `json:"proxy"`
			VPN     bool `json:"vpn"`
			Tor     bool `json:"tor"`
			Hosting bool `json:"hosting"`
		} `json:"privacy"`
	} `json:"data"`
}

// ipapiResp is the cross-check payload from ip-api.com (only the fields we need).
type ipapiResp struct {
	Status  string `json:"status"`
	Proxy   bool   `json:"proxy"`
	Hosting bool   `json:"hosting"`
}

const (
	ipinfoDemoURL = "https://ipinfo.io/widget/demo/"
	// 注意：ip-api 免费接口官方不支持 HTTPS（HTTPS 是 Pro 付费特性），
	// 在 TLS 下会不稳定地返回 403。故这里必须用 http:// 而非 https://。
	// 实测：https 返回 403，http 返回 200。
	ipapiURL = "http://ip-api.com/json/"
)

// ipClassCacheEntry caches a classification result. IP geolocation rarely
// changes, so a 3-day TTL collapses repeat node tests of the same IP into a
// single external query even across frps restarts (the cache is persisted to
// disk). Failed lookups are cached briefly (5 min) to avoid hammering the
// rate-limited endpoint.
type ipClassCacheEntry struct {
	c      ipClassInfo
	ok     bool
	expire time.Time
}

// ipClassTTL is how long a successful classification stays valid. IP geolocation
// and usage type almost never change, so 3 days is safe and keeps the
// rate-limited ipinfo/ip-api endpoints quiet.
const ipClassTTL = 72 * time.Hour

// ipClassCacheFile is the on-disk database (JSON) that mirrors ipClassCache so
// results survive frps restarts. It is written next to the frps working dir.
const ipClassCacheFile = "ip_class_cache.json"

var ipClassCache sync.Map

// ipClassPersistence guards the (debounced) background save of ipClassCache to
// disk.
var (
	ipClassSaveMu    sync.Mutex
	ipClassSaveTimer *time.Timer
	ipClassLoadOnce  sync.Once
)

// ipClassPromise represents an in-flight classification so that concurrent
// requests for the same IP wait for one query instead of each firing their own
// (which would both miss the cache and double the load on the rate-limited API).
type ipClassPromise struct {
	done chan struct{}
	c    ipClassInfo
	ok   bool
}

// ipClassInFlight tracks classifications currently being computed, keyed by IP.
var ipClassInFlight sync.Map

// ipClassCacheOnDisk mirrors ipClassCacheEntry for JSON serialization.
type ipClassCacheOnDisk struct {
	Entries map[string]struct {
		C      ipClassInfo `json:"c"`
		Ok     bool        `json:"ok"`
		Expire int64       `json:"expire"`
	} `json:"entries"`
}

// loadIPClassCache reads the on-disk database into ipClassCache at first use.
// Expired entries are silently skipped. Safe to call multiple times; the actual
// load happens only once.
func loadIPClassCache() {
	ipClassLoadOnce.Do(func() {
		data, err := os.ReadFile(ipClassCacheFile)
		if err != nil {
			// No cache file yet (first run or it was deleted) — start empty.
			return
		}
		var disk ipClassCacheOnDisk
		if err := json.Unmarshal(data, &disk); err != nil {
			return
		}
		now := time.Now()
		for ip, e := range disk.Entries {
			if now.Before(time.Unix(e.Expire, 0)) {
				ipClassCache.Store(ip, ipClassCacheEntry{c: e.C, ok: e.Ok, expire: time.Unix(e.Expire, 0)})
			}
		}
	})
}

// scheduleIPClassSave persists ipClassCache to disk on a debounce, so a burst of
// new classifications coalesces into a single write.
func scheduleIPClassSave() {
	ipClassSaveMu.Lock()
	defer ipClassSaveMu.Unlock()
	if ipClassSaveTimer != nil {
		ipClassSaveTimer.Stop()
	}
	ipClassSaveTimer = time.AfterFunc(2*time.Second, saveIPClassCache)
}

// saveIPClassCache snapshots ipClassCache to the on-disk JSON database.
func saveIPClassCache() {
	ipClassSaveMu.Lock()
	defer ipClassSaveMu.Unlock()
	disk := ipClassCacheOnDisk{Entries: make(map[string]struct {
		C      ipClassInfo `json:"c"`
		Ok     bool        `json:"ok"`
		Expire int64       `json:"expire"`
	})}
	ipClassCache.Range(func(k, v any) bool {
		e := v.(ipClassCacheEntry)
		disk.Entries[k.(string)] = struct {
			C      ipClassInfo `json:"c"`
			Ok     bool        `json:"ok"`
			Expire int64       `json:"expire"`
		}{C: e.c, Ok: e.ok, Expire: e.expire.Unix()}
		return true
	})
	data, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return
	}
	// Write to a temp file then rename, so a crash mid-write does not corrupt
	// the active database.
	tmp := ipClassCacheFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, ipClassCacheFile)
}

// classifyIP returns the classification for ip, using a process-wide cache so
// that repeated node tests (which share the same egress IPs) hit ipinfo/ip-api
// at most once per IP per 3 days. Concurrent misses for the same IP are collapsed
// into a single query via ipClassInFlight, and the result is persisted to disk
// so it survives frps restarts.
func classifyIP(ip string) (ipClassInfo, bool) {
	loadIPClassCache()
	if cached, ok := ipClassCache.Load(ip); ok {
		e := cached.(ipClassCacheEntry)
		if time.Now().Before(e.expire) {
			return e.c, e.ok
		}
	}
	// In-flight dedupe: if a classification for this IP is already running,
	// wait for it instead of launching a duplicate external query.
	if p, ok := ipClassInFlight.Load(ip); ok {
		promise := p.(*ipClassPromise)
		<-promise.done
		return promise.c, promise.ok
	}
	promise := &ipClassPromise{done: make(chan struct{})}
	actual, loaded := ipClassInFlight.LoadOrStore(ip, promise)
	if loaded {
		// Another goroutine won the race; wait for its result.
		<-actual.(*ipClassPromise).done
		return actual.(*ipClassPromise).c, actual.(*ipClassPromise).ok
	}
	// We are the leader: compute, cache, and broadcast to waiters.
	var c ipClassInfo
	var ok bool
	func() {
		defer func() {
			close(promise.done)
			ipClassInFlight.Delete(ip)
		}()
		c, _ = doClassifyIP(ip)
		ok = c.Country != "" || c.ASN != "" || c.IPType != ""
		ttl := ipClassTTL
		if !ok {
			ttl = 5 * time.Minute
		}
		ipClassCache.Store(ip, ipClassCacheEntry{c: c, ok: ok, expire: time.Now().Add(ttl)})
		scheduleIPClassSave()
		promise.c = c
		promise.ok = ok
	}()
	return c, ok
}

func doClassifyIP(ip string) (ipClassInfo, string) {
	var c ipClassInfo
	if net.ParseIP(ip) == nil {
		return c, "IP 非法: " + ip
	}
	diag := strings.Builder{}

	// ipinfo 是主数据源，ip-api 作交叉核对。两者相互独立，并发执行，
	// 把缓存未命中时的查询耗时从“串行两段”降到“取较慢的一段”。
	type ipinfoRes struct {
		wi  ipinfoWidget
		err error
	}
	type ipapiRes struct {
		ar  ipapiResp
		err error
	}
	ipinfoCh := make(chan ipinfoRes, 1)
	ipapiCh := make(chan ipapiRes, 1)

	go func() {
		body, e := httpGetNoProxy(ipinfoDemoURL+ip, 10*time.Second)
		if e != nil {
			ipinfoCh <- ipinfoRes{err: e}
			return
		}
		var wi ipinfoWidget
		e = json.Unmarshal(body, &wi)
		ipinfoCh <- ipinfoRes{wi: wi, err: e}
	}()
	go func() {
		b, e := httpGetNoProxy(ipapiURL+ip+"?fields=status,proxy,hosting", 10*time.Second)
		if e != nil {
			ipapiCh <- ipapiRes{err: e}
			return
		}
		var ar ipapiResp
		e = json.Unmarshal(b, &ar)
		ipapiCh <- ipapiRes{ar: ar, err: e}
	}()

	ir := <-ipinfoCh
	if ir.err != nil {
		return c, "ipinfo 查询失败: " + ir.err.Error()
	}
	wi := ir.wi
	c.Country = wi.Data.Country
	c.RegisteredCountry = wi.Data.Abuse.Country
	c.ASN = strings.TrimPrefix(wi.Data.ASN.ASN, "AS")
	c.Org = wi.Data.ASN.Name
	c.City = wi.Data.City
	c.Region = wi.Data.Region
	c.UsageType = wi.Data.ASN.Type
	c.CompanyType = wi.Data.Company.Type

	if c.Country != "" && c.RegisteredCountry != "" {
		if c.Country == c.RegisteredCountry {
			c.IPType, c.IsNative = "原生IP", true
		} else {
			c.IPType, c.IsBroadcast = "广播IP", true
		}
	} else {
		c.IPType = "未知"
	}

	switch c.UsageType {
	case "hosting":
		c.IsHosting = true
	case "isp":
		c.IsISP = true
	}
	if wi.Data.Privacy.Hosting {
		c.IsHosting = true
	}

	// 交叉核对 ip-api（proxy / hosting 布尔）。失败仅记录，不阻塞。
	ar := <-ipapiCh
	if ar.err == nil {
		if ar.ar.Status == "success" && ar.ar.Hosting {
			c.IsHosting = true
		}
	} else {
		diag.WriteString("ip-api 交叉核对失败(忽略): " + ar.err.Error() + "; ")
	}

	var flags []string
	if wi.Data.Privacy.Proxy {
		flags = append(flags, "代理")
	}
	if wi.Data.Privacy.VPN {
		flags = append(flags, "VPN")
	}
	if wi.Data.Privacy.Tor {
		flags = append(flags, "Tor")
	}
	if c.IsHosting {
		flags = append(flags, "机房")
	}
	if len(flags) == 0 {
		c.Privacy = "无"
	} else {
		c.Privacy = strings.Join(flags, " ")
	}

	return c, diag.String()
}

// httpGetNoProxy fetches urlStr directly, bypassing any HTTP(S)_PROXY env so the
// query is NOT routed through any proxy configured on the frps host.
func httpGetNoProxy(urlStr string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: nil,
			// 强制 HTTP/1.1：ip-api 免费接口对 Go 默认协商的 HTTP/2 很挑剔，
			// 容易返回 403；curl 默认走 HTTP/1.1 反而正常。
			TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		},
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	// 用常规 curl 风格的 UA，避免自定义 UA 被 WAF/接口判为异常客户端。
	req.Header.Set("User-Agent", "curl/8.4.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func usageLabel(t string) string {
	switch t {
	case "hosting":
		return "机房"
	case "isp":
		return "运营商"
	case "business":
		return "企业"
	case "education":
		return "教育"
	default:
		return "其他"
	}
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// applyIPClassification merges ipClassInfo fields into the node-test output map.
func applyIPClassification(data map[string]any, c ipClassInfo) {
	data["ip_country"] = c.Country
	data["ip_registered_country"] = c.RegisteredCountry
	data["ip_asn"] = c.ASN
	data["ip_org"] = c.Org
	data["ip_usage"] = usageLabel(c.UsageType)
	data["ip_type"] = c.IPType
	data["is_isp"] = boolToStr(c.IsISP)
	data["is_hosting"] = boolToStr(c.IsHosting)
	data["is_native"] = boolToStr(c.IsNative)
	data["is_broadcast"] = boolToStr(c.IsBroadcast)
	data["ip_privacy"] = c.Privacy
}
