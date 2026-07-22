// Copyright 2026 The frp Authors
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

// Package iplookup provides IP geolocation and ISP lookup via the cip.cc service.
package iplookup

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	cipCCURL     = "http://www.cip.cc/%s"
	httpTimeout  = 5 * time.Second
	cacheTTL     = 24 * time.Hour
	maxCacheSize = 10000
)

// Result holds the geolocation and ISP information for an IP address.
type Result struct {
	Location string // e.g. "中国 重庆 重庆"
	ISP      string // e.g. "联通"
}

type cacheEntry struct {
	result    Result
	fetchedAt time.Time
}

// LookupService queries cip.cc for IP geolocation info with an in-memory cache.
type LookupService struct {
	mu    sync.RWMutex
	cache map[string]cacheEntry

	client *http.Client
}

// NewLookupService creates a new IP lookup service.
func NewLookupService() *LookupService {
	return &LookupService{
		cache: make(map[string]cacheEntry),
		client: &http.Client{
			Timeout: httpTimeout,
		},
	}
}

// Lookup queries the geolocation and ISP info for the given IP address.
// Results are cached for 24 hours. Private/reserved IPs are skipped.
func (s *LookupService) Lookup(ctx context.Context, ip string) (Result, error) {
	if ip == "" {
		return Result{}, fmt.Errorf("empty ip")
	}

	// Skip private/reserved IPs
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || isPrivateIP(parsedIP) {
		return Result{}, nil
	}

	// Check cache
	s.mu.RLock()
	if entry, ok := s.cache[ip]; ok && time.Since(entry.fetchedAt) < cacheTTL {
		s.mu.RUnlock()
		return entry.result, nil
	}
	s.mu.RUnlock()

	// Query cip.cc
	result, err := s.queryCipCC(ctx, ip)
	if err != nil {
		return Result{}, err
	}

	// Store in cache
	s.mu.Lock()
	if len(s.cache) >= maxCacheSize {
		s.evictOldest()
	}
	s.cache[ip] = cacheEntry{result: result, fetchedAt: time.Now()}
	s.mu.Unlock()

	return result, nil
}

func (s *LookupService) queryCipCC(ctx context.Context, ip string) (Result, error) {
	url := fmt.Sprintf(cipCCURL, ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Result{}, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("cip.cc returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return Result{}, err
	}

	return parseCipCCResponse(string(body)), nil
}

// parseCipCCResponse parses the plain-text response from cip.cc.
// Format:
//
//	IP: 114.114.114.114
//	地址    : 中国 江苏 南京
//	运营商  : 南京信风网络科技有限公司GreatbitDNS服务器
//	数据二  : 中国 江苏 南京 | 南京信风网络科技有限公司GreatbitDNS服务器
//	数据三  : 中国 江苏省 南京市
//	URL     : http://www.cip.cc/114.114.114.114
func parseCipCCResponse(body string) Result {
	var result Result
	for line := range strings.SplitSeq(body, "\n") {
		line = strings.TrimSpace(line)
		if v, ok := extractField(line, "地址"); ok {
			result.Location = v
		} else if v, ok := extractField(line, "运营商"); ok {
			result.ISP = v
		}
	}
	return result
}

func extractField(line, prefix string) (string, bool) {
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(line, prefix)
	rest = strings.TrimSpace(rest)
	if !strings.HasPrefix(rest, ":") {
		return "", false
	}
	value := strings.TrimPrefix(rest, ":")
	return strings.TrimSpace(value), true
}

func (s *LookupService) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, v := range s.cache {
		if first || v.fetchedAt.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.fetchedAt
			first = false
		}
	}
	if oldestKey != "" {
		delete(s.cache, oldestKey)
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
		return true
	}
	privateRanges := []struct {
		network *net.IPNet
	}{
		{mustParseCIDR("10.0.0.0/8")},
		{mustParseCIDR("172.16.0.0/12")},
		{mustParseCIDR("192.168.0.0/16")},
		{mustParseCIDR("fc00::/7")},
	}
	for _, r := range privateRanges {
		if r.network.Contains(ip) {
			return true
		}
	}
	return false
}

func mustParseCIDR(s string) *net.IPNet {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return network
}
