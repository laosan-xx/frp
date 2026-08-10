package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	httppkg "github.com/laosan-xx/frp/pkg/util/http"
	"github.com/laosan-xx/frp/pkg/util/log"
)

// downloadProxy is the GitHub download proxy for Chinese clients.
// API calls go directly to api.github.com from the overseas server,
// but download URLs are rewritten to use this proxy.
const downloadProxy = "https://gh.2026178.xyz"

// firmwareCache caches GitHub API responses on the server side.
var firmwareCache struct {
	sync.Mutex
	entries map[string]*firmwareCacheEntry
}

type firmwareCacheEntry struct {
	data      string // parsed JSON response (already processed)
	etag      string // ETag for conditional requests
	expiresAt time.Time
}

const firmwareCacheTTL = 10 * time.Minute

// normalizeRepoAPI converts a proxy URL back to the real GitHub API URL.
// Old clients may send URLs through gh.2026178.xyz/api/repos/...
func normalizeRepoAPI(repoAPI string) string {
	// If it already points to api.github.com, return as-is.
	if strings.Contains(repoAPI, "api.github.com") {
		return repoAPI
	}
	// Extract the path after /api/repos/ and prepend api.github.com.
	const marker = "/api/repos/"
	idx := strings.Index(repoAPI, marker)
	if idx >= 0 {
		return "https://api.github.com" + repoAPI[idx:]
	}
	return repoAPI
}

// NewFirmwareHandler creates a firmware release handler with the given GitHub token.
// The token is read from the server config (githubToken field in frps.toml).
func NewFirmwareHandler(githubToken string) httppkg.APIHandler {
	if githubToken != "" {
		masked := githubToken
		if len(masked) > 8 {
			masked = masked[:4] + "***" + masked[len(masked)-4:]
		}
		log.Infof("firmware API proxy: GitHub token configured [%s], rate limit 5000 req/h", masked)
	} else {
		log.Warnf("firmware API proxy: no GitHub token configured, using anonymous requests (60 req/h limit)")
	}
	return func(ctx *httppkg.Context) (any, error) {
		return fetchFirmwareReleases(ctx, githubToken)
	}
}

// fetchFirmwareReleases proxies GitHub release API from the server side.
// The server (overseas) calls api.github.com directly — no shared proxy, no rate-limit issues.
// Download URLs in the response are rewritten to use the download proxy for Chinese clients.
func fetchFirmwareReleases(ctx *httppkg.Context, githubToken string) (any, error) {
	repoAPI := ctx.Query("repoApi")
	boardModel := ctx.Query("boardModel")
	if repoAPI == "" || boardModel == "" {
		return nil, httppkg.NewError(http.StatusBadRequest, "repoApi and boardModel are required")
	}

	// Normalize: old clients may send proxy URL, convert to real GitHub API URL.
	repoAPI = normalizeRepoAPI(repoAPI)

	// Check cache (fresh).
	cacheKey := repoAPI + "|" + boardModel
	firmwareCache.Lock()
	if firmwareCache.entries == nil {
		firmwareCache.entries = make(map[string]*firmwareCacheEntry)
	}
	cachedEntry := firmwareCache.entries[cacheKey]
	if cachedEntry != nil && time.Now().Before(cachedEntry.expiresAt) {
		firmwareCache.Unlock()
		return json.RawMessage(cachedEntry.data), nil
	}
	// Grab stale ETag for conditional request.
	var staleEtag string
	if cachedEntry != nil {
		staleEtag = cachedEntry.etag
	}
	firmwareCache.Unlock()

	// Build request to GitHub API (direct, no proxy).
	client := &http.Client{Timeout: 20 * time.Second}
	apiReq, err := http.NewRequest("GET", repoAPI+"/releases?per_page=30", nil)
	if err != nil {
		return nil, fmt.Errorf("构建请求失败: %v", err)
	}
	apiReq.Header.Set("User-Agent", "frp-firmware-server/1.0")
	apiReq.Header.Set("Accept", "application/vnd.github+json")
	if githubToken != "" {
		apiReq.Header.Set("Authorization", "Bearer "+githubToken)
	}
	if staleEtag != "" {
		apiReq.Header.Set("If-None-Match", staleEtag)
	}

	resp, err := client.Do(apiReq)
	if err != nil {
		// Network error – serve stale cache if available.
		if cachedEntry != nil {
			return json.RawMessage(cachedEntry.data), nil
		}
		return nil, fmt.Errorf("GitHub API 请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 304 Not Modified – refresh TTL and reuse cached data (no rate limit cost).
	if resp.StatusCode == http.StatusNotModified && cachedEntry != nil {
		firmwareCache.Lock()
		cachedEntry.expiresAt = time.Now().Add(firmwareCacheTTL)
		firmwareCache.Unlock()
		return json.RawMessage(cachedEntry.data), nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		if resp.StatusCode >= 500 && cachedEntry != nil {
			return json.RawMessage(cachedEntry.data), nil
		}
		return nil, fmt.Errorf("GitHub API 返回 %d: %s", resp.StatusCode, string(body))
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// Parse releases and filter by boardModel, rewrite download URLs to proxy.
	result := parseReleasesServer(string(rawBody), boardModel)
	// Include GitHub rate limit info in response for debugging.
	result["_rateLimit"] = map[string]string{
		"limit":     resp.Header.Get("X-RateLimit-Limit"),
		"remaining": resp.Header.Get("X-RateLimit-Remaining"),
		"reset":     resp.Header.Get("X-RateLimit-Reset"),
		"token":     boolStr(githubToken != ""),
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("JSON 序列化失败: %v", err)
	}

	// Cache the parsed result.
	firmwareCache.Lock()
	firmwareCache.entries[cacheKey] = &firmwareCacheEntry{
		data:      string(jsonBytes),
		etag:      resp.Header.Get("ETag"),
		expiresAt: time.Now().Add(firmwareCacheTTL),
	}
	firmwareCache.Unlock()

	return json.RawMessage(jsonBytes), nil
}

// parseReleasesServer decodes GitHub releases JSON, filters by board model,
// and rewrites download URLs to use the download proxy.
func parseReleasesServer(jsonData string, boardModel string) map[string]any {
	var ghReleases []struct {
		Name   string `json:"name"`
		Assets []struct {
			Name               string `json:"name"`
			Size               int64  `json:"size"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.Unmarshal([]byte(jsonData), &ghReleases); err != nil {
		return map[string]any{"error": fmt.Sprintf("解析 GitHub 响应失败: %v", err)}
	}

	type branchEntry struct {
		Branch string
		Config string
		Date   string
		Assets []map[string]any
	}
	groups := make(map[string]*branchEntry)

	for _, rel := range ghReleases {
		config, branch, date := parseReleaseName(rel.Name)
		if config == "" || branch == "" {
			continue
		}

		// Filter assets: name contains boardModel, size > 30MB, and exclude factory images.
		var matchedAssets []map[string]any
		for _, a := range rel.Assets {
			lowerName := strings.ToLower(a.Name)
			if strings.Contains(lowerName, "factory") {
				continue
			}
			if strings.Contains(a.Name, boardModel) && a.Size > 30*1024*1024 {
				// Rewrite download URL to use download proxy.
				dlURL := strings.Replace(a.BrowserDownloadURL, "https://github.com/", downloadProxy+"/", 1)
				matchedAssets = append(matchedAssets, map[string]any{
					"name": a.Name,
					"size": a.Size,
					"url":  dlURL,
				})
			}
		}
		if len(matchedAssets) == 0 {
			continue
		}

		key := config + "|" + branch
		existing, ok := groups[key]
		if !ok || date > existing.Date {
			groups[key] = &branchEntry{
				Branch: branch,
				Config: config,
				Date:   date,
				Assets: matchedAssets,
			}
		}
	}

	branches := make([]map[string]any, 0, len(groups))
	for _, entry := range groups {
		branches = append(branches, map[string]any{
			"branch": entry.Branch,
			"config": entry.Config,
			"date":   entry.Date,
			"assets": entry.Assets,
		})
	}

	return map[string]any{"branches": branches}
}

// parseReleaseName parses release names like:
// IPQ60XX-WIFI-YES-LAOSAN-second-26.07.28-03.02.24
// Returns (config, branch, date)
func parseReleaseName(name string) (config, branch, date string) {
	parts := strings.SplitN(name, "-LAOSAN-", 2)
	if len(parts) != 2 {
		return "", "", ""
	}
	config = strings.TrimSpace(parts[0])
	after := strings.TrimSpace(parts[1])

	// The last two dash-separated segments are date: YY.MM.DD-HH.MM.SS
	// Everything before is branch
	// e.g. "second-26.07.28-03.02.24" -> branch="second", date="26.07.28-03.02.24"
	dashIdx := strings.LastIndex(after, "-")
	if dashIdx < 0 {
		return config, after, ""
	}
	secondLast := after[:dashIdx]
	dateStr := after[dashIdx+1:]

	dashIdx2 := strings.LastIndex(secondLast, "-")
	if dashIdx2 < 0 {
		return config, after, ""
	}
	branch = secondLast[:dashIdx2]
	datePart := secondLast[dashIdx2+1:]
	date = datePart + "-" + dateStr
	return config, branch, date
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
