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

// Package iplookup provides IP geolocation and ISP lookup using a local ip2region xdb database.
package iplookup

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
)

//go:embed ip2region_v4.xdb
var ip2regionData []byte

// Result holds the geolocation and ISP information for an IP address.
type Result struct {
	Location string // e.g. "中国 江苏 南京"
	ISP      string // e.g. "电信"
}

// LookupService queries the local ip2region xdb database for IP geolocation info.
type LookupService struct {
	mu       sync.Mutex
	searcher *xdb.Searcher
}

// NewLookupService creates a new IP lookup service backed by the embedded ip2region database.
func NewLookupService() *LookupService {
	searcher, err := xdb.NewWithBuffer(xdb.IPv4, ip2regionData)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize ip2region searcher: %v", err))
	}
	return &LookupService{
		searcher: searcher,
	}
}

// Lookup returns the geolocation and ISP info for the given IP address.
// Private/reserved IPs return an empty result.
func (s *LookupService) Lookup(_ context.Context, ip string) (Result, error) {
	if ip == "" {
		return Result{}, fmt.Errorf("empty ip")
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || isPrivateIP(parsedIP) {
		return Result{}, nil
	}

	// Only support IPv4
	if parsedIP.To4() == nil {
		return Result{}, nil
	}

	// Searcher is not thread-safe, protect with mutex
	s.mu.Lock()
	region, err := s.searcher.Search(ip)
	s.mu.Unlock()
	if err != nil {
		return Result{}, fmt.Errorf("ip2region search for %s: %w", ip, err)
	}

	return parseRegion(region), nil
}

// parseRegion parses the ip2region result string.
// Format: "国家|省份|城市|ISP|国家代码"
// Example: "中国|江苏省|南京市|电信|CN"
func parseRegion(region string) Result {
	if region == "" {
		return Result{}
	}

	parts := strings.Split(region, "|")
	if len(parts) < 5 {
		return Result{}
	}

	country := parts[0]
	province := parts[1]
	city := parts[2]
	isp := parts[3]

	// Replace "0" with empty (ip2region uses "0" for unknown fields)
	if province == "0" {
		province = ""
	}
	if city == "0" {
		city = ""
	}
	if isp == "0" {
		isp = ""
	}

	// Build location string: "国家 省份 城市" (deduplicate consecutive identical parts)
	var locationParts []string
	if country != "" && country != "0" {
		locationParts = append(locationParts, country)
	}
	if province != "" && (len(locationParts) == 0 || locationParts[len(locationParts)-1] != province) {
		locationParts = append(locationParts, province)
	}
	if city != "" && (len(locationParts) == 0 || locationParts[len(locationParts)-1] != city) {
		locationParts = append(locationParts, city)
	}

	return Result{
		Location: strings.Join(locationParts, " "),
		ISP:      isp,
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
