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

package client

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/laosan-xx/frp/pkg/util/log"
)

// builtinCommandExecutor implements CommandExecutor with built-in command handlers.
// It replaces the external frpc-command-handler.sh script by handling commands
// directly within the frpc binary. All node operations use uci directly,
// no dependency on add_node.sh / set_node.sh.
type builtinCommandExecutor struct{}

func (e *builtinCommandExecutor) Execute(command, payload string) (result, output string) {
	log.Tracef("builtin command executor: handling command [%s]", command)

	switch command {
	case "node_link":
		return e.cmdNodeLink(payload)
	case "node_export":
		return "ok", nodeExport(payload)
	case "get_nodes":
		return e.cmdGetNodes()
	case "del_node":
		return e.cmdDelNode(payload)
	case "set_node":
		return e.cmdSetNode(payload)
	case "url_test_node":
		return e.cmdURLTestNode(payload)
	case "url_test_node_noiip":
		return e.cmdURLTestNodeNoIP(payload)
	case "url_test_device":
		return e.cmdURLTestDevice(payload)
	case "disable_passwall":
		return e.cmdDisablePasswall()
	case "modify_frp":
		return e.cmdModifyFrp(payload)
	case "modify_system":
		return e.cmdModifySystem(payload)
	case "get_system":
		return e.cmdGetSystem()
	case "get_default_password":
		return e.cmdGetDefaultPassword()
	case "set_default_password":
		return e.cmdSetDefaultPassword(payload)
	case "detect_platform":
		return e.cmdDetectPlatform()
	case "download_firmware":
		return e.cmdDownloadFirmware(payload)
	case "get_system_version":
		return e.cmdGetSystemVersion()
	case "download_status":
		return e.cmdDownloadStatus()
	case "cancel_download":
		return e.cmdCancelDownload()
	case "run_sysupgrade":
		return e.cmdRunSysupgrade(payload)
	default:
		return "error", fmt.Sprintf("未知命令: %s", command)
	}
}

// cmdNodeLink parses a share link and adds a passwall node via uci.
// Supported: ss:// vmess:// vless:// trojan://
func (e *builtinCommandExecutor) cmdNodeLink(payload string) (string, string) {
	if payload == "" {
		return "error", "错误: 请输入代理链接"
	}
	payload = strings.TrimSpace(payload)

	node, err := parseShareLink(payload)
	if err != nil {
		return "error", fmt.Sprintf("解析链接失败: %v", err)
	}

	// Create a named section with type "nodes" (same as passwall LuCI)
	secName := fmt.Sprintf("%x", time.Now().UnixNano())
	uciSet("passwall."+secName, "nodes")

	// Set node fields
	uciSet("passwall."+secName+".remarks", node.Remarks)
	uciSet("passwall."+secName+".type", node.Type)
	uciSet("passwall."+secName+".protocol", node.Protocol)
	uciSet("passwall."+secName+".address", node.Address)
	uciSet("passwall."+secName+".port", node.Port)

	for k, v := range node.Extra {
		uciSet("passwall."+secName+"."+k, v)
	}

	uciCommit("passwall")

	// Verify node section was created
	verifyType := uciGet("passwall." + secName)
	if verifyType != "nodes" {
		return "error", fmt.Sprintf("错误: 节点创建失败 (section=%s, type=%s)", secName, verifyType)
	}

	return "ok", fmt.Sprintf("已添加节点: %s (%s %s | %s:%s)", node.Remarks, node.Type, node.Protocol, node.Address, node.Port)
}

// nodeInfo represents a single passwall node for JSON output.
type nodeInfo struct {
	Remarks string `json:"remarks"`
	Type    string `json:"type"`
	Address string `json:"address"`
	Port    string `json:"port"`
	Active  bool   `json:"active"`
}

// cmdGetNodes lists all available proxy nodes via uci, returns JSON.
func (e *builtinCommandExecutor) cmdGetNodes() (string, string) {
	output, err := runCommand("uci", "-q", "show", "passwall")
	if err != nil {
		return "error", fmt.Sprintf("uci 查询失败: %v", err)
	}

	tcpNode := uciGet("passwall.@global[0].tcp_node")
	enabled := uciGet("passwall.@global[0].enabled")
	if enabled == "" {
		enabled = "0"
	}

	// Pass 1: collect section IDs that are type "nodes" (passwall.SECTION=nodes)
	// Use a slice to preserve config file order (same as OpenWrt Passwall LuCI)
	var nodeSecIDs []string
	seen := map[string]bool{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "=nodes") {
			secID := strings.TrimSuffix(strings.TrimPrefix(line, "passwall."), "=nodes")
			if secID != "" && secID != "@global[0]" && !seen[secID] {
				seen[secID] = true
				nodeSecIDs = append(nodeSecIDs, secID)
			}
		}
	}

	// Pass 2: get details for each node section (in config file order)
	nodes := []nodeInfo{}
	activeRemarks := ""
	for _, secID := range nodeSecIDs {
		remarks := uciGet("passwall." + secID + ".remarks")
		if remarks == "" {
			continue
		}
		active := secID == tcpNode
		if active {
			activeRemarks = remarks
		}
		nodes = append(nodes, nodeInfo{
			Remarks: remarks,
			Type:    uciGet("passwall." + secID + ".type"),
			Address: uciGet("passwall." + secID + ".address"),
			Port:    uciGet("passwall." + secID + ".port"),
			Active:  active,
		})
	}

	// Fallback: if no active node matched by secID, try matching tcpNode value by remarks
	if activeRemarks == "" && tcpNode != "" {
		for i := range nodes {
			if nodes[i].Remarks == tcpNode {
				nodes[i].Active = true
				activeRemarks = nodes[i].Remarks
				break
			}
		}
	}

	result := map[string]any{
		"nodes":      nodes,
		"enabled":    enabled == "1",
		"running":    passwallIsRunning(),
		"activeNode": activeRemarks,
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "error", fmt.Sprintf("JSON 序列化失败: %v", err)
	}
	return "ok", string(jsonBytes)
}

// cmdSetNode sets the active TCP node by remarks name and restarts passwall.
func (e *builtinCommandExecutor) cmdSetNode(payload string) (string, string) {
	if payload == "" {
		return "error", "错误: 请输入节点备注名"
	}

	secID := findNodeByRemarks(payload)
	if secID == "" {
		// Detailed diagnostics using EXACT same logic as findNodeByRemarks
		output, _ := runCommand("uci", "-q", "show", "passwall")
		var avail []string
		var debugSecID, debugType string
		// Build secTypes map (same as findNodeByRemarks Pass 1)
		secTypes := map[string]string{}
		for _, line := range strings.Split(output, "\n") {
			line = strings.TrimSpace(line)
			if !strings.HasPrefix(line, "passwall.") {
				continue
			}
			rest := strings.TrimPrefix(line, "passwall.")
			eqIdx := strings.Index(rest, "=")
			if eqIdx < 0 {
				continue
			}
			sid := rest[:eqIdx]
			val := rest[eqIdx+1:]
			if !strings.Contains(sid, ".") {
				secTypes[sid] = val
			}
		}
		// Find remarks match (same as findNodeByRemarks Pass 2)
		for _, line := range strings.Split(output, "\n") {
			if !strings.Contains(line, ".remarks=") {
				continue
			}
			eqIdx := strings.Index(line, "=")
			if eqIdx < 0 {
				continue
			}
			key := line[:eqIdx]
			val := strings.Trim(line[eqIdx+1:], "'")
			avail = append(avail, val)
			if val == payload {
				// key format: "passwall.<secID>.remarks"
				keyParts := strings.SplitN(key, ".", 3)
				if len(keyParts) >= 3 {
					debugSecID = keyParts[1]
					debugType = secTypes[debugSecID]
				}
			}
		}
		return "error", fmt.Sprintf("错误: 找不到节点 '%s' (secID=%s, type='%s')，可用: %v", payload, debugSecID, debugType, avail)
	}

	uciSet("passwall.@global[0].tcp_node", secID)
	uciSet("passwall.@global[0].enabled", "1")
	uciCommit("passwall")

	// Verify the value was actually written
	verify := uciGet("passwall.@global[0].tcp_node")
	if verify != secID {
		return "error", fmt.Sprintf("错误: 设置失败，期望=%s 实际=%s (secID=%s)", secID, verify, secID)
	}

	_, _ = runCommand("/etc/init.d/passwall", "restart")

	return "ok", fmt.Sprintf("已启用节点: %s (section=%s)", payload, secID)
}

// cmdDelNode deletes a node by its remarks name.
func (e *builtinCommandExecutor) cmdDelNode(payload string) (string, string) {
	if payload == "" {
		return "error", "错误: 请输入要删除的节点名称"
	}

	target := findNodeByRemarks(payload)
	if target == "" {
		return "error", fmt.Sprintf("错误: 找不到节点 '%s'", payload)
	}

	var sb strings.Builder

	// Check if the node is currently in use
	tcpNode := uciGet("passwall.@global[0].tcp_node")
	if tcpNode == target {
		sb.WriteString("提示: 该节点当前正在使用，将自动停止服务\n")
		uciSet("passwall.@global[0].enabled", "0")
		// 关键修复：被删节点曾是选中的主节点时，必须同时清空 tcp_node。
		// 否则 tcp_node 仍指向已删除的 section，passwall 进入 “Not set” 损坏态，
		// 后续所有 url_test_node 都会失败，直到在 OpenWrt 后台手动置空才恢复。
		uciSet("passwall.@global[0].tcp_node", "")
		_, _ = runCommand("/etc/init.d/passwall", "stop")
	}

	// Delete the node section
	uciDelete("passwall." + target)
	uciCommit("passwall")

	// If no nodes remain, ensure passwall is disabled and selection cleared
	if !hasRemainingNodes() {
		uciSet("passwall.@global[0].enabled", "0")
		uciSet("passwall.@global[0].tcp_node", "")
		uciCommit("passwall")
		_, _ = runCommand("/etc/init.d/passwall", "stop")
	}

	sb.WriteString(fmt.Sprintf("已删除节点: %s", payload))
	return "ok", sb.String()
}

// cmdDisablePasswall stops passwall and sets enabled=0.
func (e *builtinCommandExecutor) cmdDisablePasswall() (string, string) {
	uciSet("passwall.@global[0].enabled", "0")
	// 关键修复：停用时一并清空 TCP 节点选择。否则 tcp_node 仍指向刚停用的节点，
	// passwall 进入 “Not set” 损坏态，后续所有 url_test_node 都会失败，
	// 直到在 OpenWrt 后台手动置空才恢复。
	uciSet("passwall.@global[0].tcp_node", "")
	uciCommit("passwall")
	_, _ = runCommand("/etc/init.d/passwall", "stop")
	return "ok", "已关闭 Passwall"
}

// cmdURLTestNode tests a node's connectivity (latency) and classifies its egress
// IP (原生/广播/机房/ISP). payload is the node's remarks name.
func (e *builtinCommandExecutor) cmdURLTestNode(payload string) (string, string) {
	return e.testNode(payload, false)
}

// cmdURLTestNodeNoIP is the same connectivity test but skips the egress-IP
// classification (no ip-api query). Used for testing non-current nodes where the
// IP info is irrelevant, to avoid hammering the rate-limited ip-api endpoint.
func (e *builtinCommandExecutor) cmdURLTestNodeNoIP(payload string) (string, string) {
	return e.testNode(payload, true)
}

// cmdURLTestDevice classifies a given IP (typically the device's own public IP,
// which the server already knows as the frpc connection source) without any proxy
// probe. It returns the same ip_country/ip_type/is_isp fields as url_test_node so
// the "current node info" panel can show the device IP + its geolocation/type when
// passwall is NOT running. The server side then enriches location/isp from its
// local IP database (same path as url_test_node).
func (e *builtinCommandExecutor) cmdURLTestDevice(payload string) (string, string) {
	ip := strings.TrimSpace(payload)
	if ip == "" {
		return "error", "错误: 缺少设备 IP"
	}
	resp := map[string]string{"code": "0", "ip": ip}
	jsonBytes, _ := json.Marshal(resp)
	return "ok", string(jsonBytes)
}

// testNode implements the node connectivity test. When skipIP is true, only the
// HTTP code + latency + raw egress IP are returned; the IP classification step
// (ip-api query) is skipped.
func (e *builtinCommandExecutor) testNode(payload string, skipIP bool) (string, string) {
	if payload == "" {
		return "error", "错误: 请指定节点备注名"
	}
	remarks := strings.TrimSpace(payload)

	secID := findNodeByRemarks(remarks)
	if secID == "" {
		return "error", fmt.Sprintf("错误: 找不到节点 '%s'", remarks)
	}

	nodeType := strings.ToLower(strings.TrimSpace(uciGet(fmt.Sprintf("passwall.%s.type", secID))))

	// socks 类型节点：直接复用节点自身 socks 出口做探测，无需起临时代理。
	if nodeType == "socks" {
		addr := uciGet(fmt.Sprintf("passwall.%s.address", secID))
		p := uciGet(fmt.Sprintf("passwall.%s.port", secID))
		user := uciGet(fmt.Sprintf("passwall.%s.username", secID))
		pass := uciGet(fmt.Sprintf("passwall.%s.password", secID))
		if addr == "" || p == "" {
			return "error", "错误: socks 节点缺少 address/port"
		}
		socksProxy := fmt.Sprintf("socks5h://%s:%s", addr, p)
		if user != "" && pass != "" {
			socksProxy = fmt.Sprintf("socks5h://%s:%s@%s:%s", user, pass, addr, p)
		}
		code, latency, ip := probeThroughProxy(socksProxy)
		resp := map[string]string{"code": code, "latency": latency, "ip": ip}
		jsonBytes, _ := json.Marshal(resp)
		return "ok", string(jsonBytes)
	}

	// 其它类型（Xray / Sing-Box / Trojan / V2Ray / SS 等）：
	// ① 先起一个常驻临时 SOCKS 代理（参数与官方 run_socks 完全一致），对其做多次 HTTP
	//    探测取最小延迟（剔除冷启动握手的一次性开销，更接近 v2rayN 的稳态链路口径）。
	//    代理成功 → 直接用它的结果（更快、稳态值）。
	// ② 代理起不来/超时时，用官方 test.sh url_test_node 兜底（它也会冷启动 xray 但
	//    至少保证有结果，不会让按钮直接"测试失败"）。
	// 关键在于：代理优先、官方兜底。这样既保留了提速（大多数节点走①），又不会在
	// 代理偶发失败时裸奔成 000。出口 IP 探测仅在代理成功时做（skipIP=true 跳过）。
	//
	// 关键认知：run_socks 的退出码不可信！app.sh 最后一行是 `[ "$type" != "xray" ]`
	// 这类判断，对 xray 节点求值为假 → 整个脚本 exit 1，但 xray 其实已成功启动。
	// 故：不以退出码论成败，一律等端口监听——端口开了即代理已起。

	// passwall 未运行时 /tmp/etc/passwall(bin) 可能已被清空，而 ln_run 依赖该目录给 xray
	// 建符号链接，缺失会导致 xray 报“没有执行权限”且永远起不来。先确保目录存在。
	_, _ = runCommand("sh", "-c", "mkdir -p /tmp/etc/passwall /tmp/etc/passwall/bin /tmp/etc/passwall2 /tmp/etc/passwall2/bin 2>/dev/null")

	code, latency := "000", "0"
	ip, ipErr := "", ""
	if p, cleanup, perr := startEgressProbeProxy(secID, 10*time.Second); perr == nil {
		defer cleanup()
		socksProxy := fmt.Sprintf("socks5h://127.0.0.1:%d", p)
		mCode, mLat := measureLatency(socksProxy, 2)
		if mCode != "000" && mLat != "0" {
			code, latency = mCode, mLat
		}
		if !skipIP {
			for i := 0; i < 2; i++ {
				if ip, ipErr = detectEgressIPVerbose(socksProxy); ip != "" {
					ipErr = ""
					break
				}
				time.Sleep(1500 * time.Millisecond)
			}
		}
	} else {
		// 代理起不来：回退官方 url_test_node（仍会冷启动 xray，但保证有结果）。
		if ts := passwallTestScript(); ts != "" {
			tout, _ := runProxyCmd("sh", "-c", fmt.Sprintf(
				"CONFIG=passwall NO_REC_PROCESS=1 %s url_test_node %s 2>&1", ts, secID))
			if c, l := parseTestScriptOutput(strings.TrimSpace(tout)); c != "" {
				code, latency = c, l
			}
		}
	}

	resp := map[string]string{"code": code, "latency": latency, "ip": ip}
	if ip == "" && ipErr != "" {
		resp["ip_err"] = ipErr
	}
	jsonBytes, _ := json.Marshal(resp)
	return "ok", string(jsonBytes)
}

// cmdURLTestNodeDiag is a self-discovery diagnostic. It dumps the ACTUAL source of
// run_socks() and url_test_node() from the installed passwall scripts, captures the
// raw stdout of the official test.sh, and runs our own run_socks under `sh -x`.
// detectEgressIP returns the public egress IP seen through the given socks proxy.
func detectEgressIP(socksProxy string) string {
	ip, _ := detectEgressIPVerbose(socksProxy)
	return ip
}

// detectEgressIPVerbose is detectEgressIP with a debug trail. It first fires a
// warm-up request to google.com/generate_204 (proven reachable through the very
// same proxy by passwall's official test.sh, which returns 204) so the upstream
// tunnel is fully established before we query the IP-echo services. When it
// cannot obtain an IP, the returned debug string records exactly which step
// failed (warm-up error, or each service's curl err / raw output) so the caller
// can surface it instead of failing silently.
//
// Note: we use the bare `curl` (resolved via PATH) exactly like test.sh does —
// that combination is proven to work on the router, so it is the reference.
func detectEgressIPVerbose(socksProxy string) (ip, debug string) {
	var sb strings.Builder

	// 1) warm-up：先打通上游隧道（官方 test.sh 已证明该地址经此代理可达，返回 204）。
	// 超时收紧到 4s，使整个节点测速能在服务端的 10s 上限内返回。
	warm := fmt.Sprintf("curl -x '%s' -o /dev/null -s -m 4 https://www.google.com/generate_204", socksProxy)
	if _, werr := runCommand("sh", "-c", warm); werr != nil {
		sb.WriteString("warmup(google) err=" + werr.Error() + "; ")
	} else {
		sb.WriteString("warmup(google) ok; ")
	}

	// 2) 逐个 IP 回显服务探测，返回第一个合法 IP。
	candidates := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://myip.ipip.net",
		"https://ipinfo.io/ip",
	}
	for _, u := range candidates {
		cmd := fmt.Sprintf("curl -x '%s' -4 --connect-timeout 3 --max-time 4 -s '%s'", socksProxy, u)
		out, err := runCommand("sh", "-c", cmd)
		if err != nil {
			sb.WriteString(u + " err=" + err.Error() + "; ")
			continue
		}
		val := strings.TrimSpace(out)
		// myip.ipip.net 返回 "IP 归属地" 文本，取首个 token。
		if fields := strings.Fields(val); len(fields) > 0 {
			val = fields[0]
		}
		if net.ParseIP(val) != nil {
			return val, sb.String()
		}
		sb.WriteString(u + " out=" + strconv.Quote(val) + "; ")
	}
	return "", sb.String()
}

// waitPortOpen polls tcp addr until it accepts a connection or the timeout elapses.
// Used to confirm a freshly-launched passwall proxy is actually ready to serve
// before we push an egress-IP probe through it (Trojan/VLESS/TLS nodes often
// need a few seconds to establish upstream).
func waitPortOpen(addr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

// passwallTestScript returns the path of the installed passwall test script.
// v1 lives at /usr/share/passwall/test.sh, v2 at /usr/share/passwall2/test.sh.
func passwallTestScript() string {
	for _, p := range []string{"/usr/share/passwall/test.sh", "/usr/share/passwall2/test.sh"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// passwallAppScript returns the matching app.sh next to the detected test.sh,
// so we invoke run_socks from the same passwall version (v1 or v2).
func passwallAppScript() string {
	if ts := passwallTestScript(); ts != "" {
		return strings.TrimSuffix(ts, "test.sh") + "app.sh"
	}
	return "/usr/share/passwall/app.sh"
}

// runSocksCmd builds the run_socks invocation EXACTLY the way passwall's own
// test.sh does: `NO_REC_PROCESS=1 /usr/share/<CONFIG>/app.sh run_socks ...`.
// Note test.sh does NOT pass CONFIG as an environment variable — it only uses
// $CONFIG to build the path, and app.sh self-determines CONFIG. We must not
// prefix `CONFIG=passwall` either: empirically that extra env var is what made
// run_socks exit 1 (the working sh -x trace in the diag omitted it).
func runSocksCmd(flag, secID string, port int) string {
	return fmt.Sprintf(
		"NO_REC_PROCESS=1 %s run_socks flag=\"%s\" node=%s bind=127.0.0.1 socks_port=%d config_file=%s.json",
		passwallAppScript(), flag, secID, port, flag,
	)
}

// traceRunSocks runs the very same run_socks invocation under `sh -x` and
// returns the tail of the trace, so a silent `exit status 1` becomes debuggable
// (app.sh logs to syslog, leaving stdout empty; the sh -x trace shows the
// actual failing line/variable).
func traceRunSocks(flag, secID string, port int) string {
	trace := fmt.Sprintf(
		"NO_REC_PROCESS=1 sh -x %s run_socks flag=\"%s\" node=%s bind=127.0.0.1 socks_port=%d config_file=%s.json 2>&1 | tail -40",
		passwallAppScript(), flag, secID, port, flag,
	)
	out, _ := runProxyCmd("sh", "-c", trace)
	return "\n--- run_socks sh -x (tail) ---\n" + out
}

var (
	reHTTPCode  = regexp.MustCompile(`\b[1-5][0-9]{2}\b`)
	reLatencyMS = regexp.MustCompile(`([0-9]+(?:\.[0-9]+)?)\s*ms`)
	// 官方 test.sh url_test_node 输出格式："HTTP码:延迟秒数"，如 "204:1.694349"
	reTestOut = regexp.MustCompile(`\b([0-9]{3}):([0-9]+(?:\.[0-9]+)?)\b`)
)

// parseTestScriptOutput parses passwall's url_test_node stdout. The canonical
// output format is "HTTPCODE:SECONDS" (e.g. "204:1.694349"); we parse that
// directly. Older/other formats that print a bare 3-digit code plus an optional
// "<num>ms" are also handled. Used as the fallback when the temporary proxy
// fails to start.
func parseTestScriptOutput(out string) (code, latency string) {
	out = strings.TrimSpace(out)
	if m := reTestOut.FindStringSubmatch(out); m != nil {
		return m[1], m[2]
	}
	// legacy fallback
	code = "200"
	latency = "0"
	if m := reHTTPCode.FindString(out); m != "" {
		code = m
	}
	if m := reLatencyMS.FindStringSubmatch(out); m != nil {
		latency = m[1]
	}
	return code, latency
}

// probeThroughProxy runs a latency + egress-IP probe through the given socks proxy
// and returns the HTTP code, pretransfer latency (seconds) and the public egress IP.
func probeThroughProxy(socksProxy string) (code, latency, ip string) {
	probeURL := "https://www.google.com/generate_204"
	latencyCmd := fmt.Sprintf(
		"/usr/bin/curl -x '%s' --connect-timeout 3 --max-time 5 -o /dev/null -I -skL -w '%%{http_code}:%%{time_pretransfer}' '%s'",
		socksProxy, probeURL,
	)
	latOut, _ := runCommand("sh", "-c", latencyCmd)
	latResult := strings.TrimSpace(latOut)
	code = "000"
	latency = "0"
	if parts := strings.SplitN(latResult, ":", 2); len(parts) == 2 {
		code = parts[0]
		latency = parts[1]
	}
	ip = detectEgressIP(socksProxy)
	return code, latency, ip
}

// measureLatency probes the node through an already-running local proxy several
// times and returns the HTTP code plus the *best* (smallest) latency in seconds.
//
// Why multiple probes + take minimum:
//   - The proxy was just (cold-)started, so the very first request usually pays
//     the upstream TLS handshake + tunnel warm-up cost. Later probes measure the
//     steady-state link latency, which is what users actually care about and what
//     desktop clients (e.g. v2rayN) approximate.
//   - A single probe is vulnerable to one-off network jitter, inflating the number.
//
// time_pretransfer is used because it captures "up to the moment the server
// response starts" (proxy -> node -> server handshake), i.e. the node latency,
// without counting the (large) response body download.
func measureLatency(socksProxy string, times int) (code, bestSeconds string) {
	probeURL := "https://www.google.com/generate_204"
	// 必须用 socks5h://（让代理端 xray 做 DNS 解析），与官方 url_test_node 完全一致。
	// 若用 socks5:// 则由本地 curl 解析域名，而路由器本地 DNS 往往解析不了
	// google.com（被污染/无海外 DNS）→ CONNECT 到错误 IP → TLS 握手超时（code 000）。
	// run_socks 生成的 xray 配置了 direct_dns(223.5.5.5)，由它解析才可达。
	localProxy := strings.Replace(socksProxy, "socks5://", "socks5h://", 1)
	lastCode := "000"
	best := -1.0
	// warmup：先打通一次上游隧道。xray 刚起来时首次握手含额外开销，且部分节点首次
	// CONNECT 需要建立上游连接；先预热可避免后面每次都付这笔一次性成本，也能确认代理
	// 链路本身可达（不可达则后面也全是 000，但至少不被误判为"有值"）。
	warm := fmt.Sprintf(
		"/usr/bin/curl -x '%s' --connect-timeout 3 --max-time 5 -o /dev/null -s -k -w '%%{http_code}' '%s'",
		localProxy, probeURL,
	)
	if wout, werr := runCommand("sh", "-c", warm); werr == nil {
		lastCode = strings.TrimSpace(wout)
	}
	for i := 0; i < times; i++ {
		cmd := fmt.Sprintf(
			"/usr/bin/curl -x '%s' --connect-timeout 3 --max-time 4 -o /dev/null -s -k -w '%%{http_code}:%%{time_pretransfer}' '%s'",
			localProxy, probeURL,
		)
		out, _ := runCommand("sh", "-c", cmd)
		out = strings.TrimSpace(out)
		parts := strings.SplitN(out, ":", 2)
		if len(parts) != 2 {
			continue
		}
		lastCode = parts[0]
		// 视为可达的码：204/200/000(实际有返回但 curl 视连接成功) 之外的正常响应。
		if parts[0] == "000" {
			continue
		}
		t, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}
		if best < 0 || t < best {
			best = t
		}
	}
	if best < 0 {
		return lastCode, "0"
	}
	return lastCode, strconv.FormatFloat(best, 'f', 3, 64)
}

// IP 分类（原生/广播/机房/住宅、国家、ASN、隐私标记）已迁移到 frps 侧
// （server/http/ipclassify.go）：frpc 只负责探测并回传裸出口 IP，frps 统一
// 向 ipinfo.io / ip-api 查询并缓存，避免每台路由器各自打满限流额度。

// getFreePort asks the OS for a currently-free local TCP port.
func getFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// tempProxyMu serializes the temporary-proxy port reservation handshake so that
// two concurrent url_test_node clicks cannot grab the same local port (the
// TOCTOU race that previously made one of two simultaneous tests fail with
// "代理未在10s内监听").
var tempProxyMu sync.Mutex

// startEgressProbeProxy reserves a unique local port, launches a passwall
// temporary proxy on it via run_socks, waits until it is actually listening,
// and returns the port plus a cleanup function. The reservation listener is
// held open until the proxy confirms listening, which guarantees the port is
// never free for another concurrent test to grab in between.
func startEgressProbeProxy(secID string, timeout time.Duration) (port int, cleanup func(), err error) {
	tempProxyMu.Lock()
	// 仅借 getFreePort 选一个当前空闲的端口号，拿到后立即释放监听（不持有该端口），
	// 否则 Go 持有了端口、xray 再 bind 同一端口会失败（Address already in use）→
	// xray 实际没监听，waitPortOpen 却检测到的是 Go 自己的监听，ln.Close() 后端口
	// 彻底空掉，curl 连接被拒（exit status 7 / code 000）。
	fp, ferr := getFreePort()
	if ferr != nil {
		tempProxyMu.Unlock()
		return 0, nil, ferr
	}
	port = fp
	// 必须用官方标准前缀 url_test_<secID>：passwall run_socks 是按 flag 前缀识别
	// "这是测速、需生成完整节点出站配置" 的。之前用 url_test_ip_ 前缀 → 生成的 xray
	// 配置 outbound 为空（直连），TLS Client Hello 发到公网后无响应，curl 全超时 000。
	// 不同节点 secID 不同 → flag 天然不冲突；同节点重复点击由 cleanup 先清解决。
	ipFlag := "url_test_" + secID
	cleanupProxyByFlag(ipFlag) // 清掉上次可能残留的同名代理/配置
	// run_socks 退出码不可信（见 testNode 注释），且它会以前台方式阻塞，故放入
	// goroutine 异步启动，主流程只等端口真正被 xray 监听即可。
	go runProxyCmd("sh", "-c", runSocksCmd(ipFlag, secID, port))
	if !waitPortOpen(fmt.Sprintf("127.0.0.1:%d", port), timeout) {
		tempProxyMu.Unlock()
		return 0, nil, fmt.Errorf("代理未在%v内监听 127.0.0.1:%d%s", timeout, port, traceRunSocks(ipFlag, secID, port))
	}
	tempProxyMu.Unlock()
	cleanup = func() { cleanupProxyByFlag(ipFlag) }
	return port, cleanup, nil
}

// cleanupTempProxy kills the temporary passwall proxy and removes its temp files.
func cleanupTempProxy(secID string) {
	cleanCmd := fmt.Sprintf(
		"busybox pgrep -af 'url_test_%s' | awk '! /test\\.sh/{print $1}' | xargs kill -9 >/dev/null 2>&1; rm -rf /tmp/etc/passwall/*urltest_%s* /tmp/etc/passwall/*url_test_%s* 2>/dev/null",
		secID, secID, secID,
	)
	_, _ = runCommand("sh", "-c", cleanCmd)
}

// cleanupProxyByFlag kills any passwall proxy started with the given run_socks
// flag and removes its generated config under /tmp/etc/passwall. Used for our
// dedicated egress-IP probe proxy (flag url_test_ip_<secID>) so it never
// collides with the official test.sh proxy (flag url_test_<secID>).
func cleanupProxyByFlag(flag string) {
	cleanCmd := fmt.Sprintf(
		"busybox pgrep -af '%s' | awk '! /test\\.sh/{print $1}' | xargs kill -9 >/dev/null 2>&1; rm -rf /tmp/etc/passwall/*%s* 2>/dev/null",
		flag, flag,
	)
	_, _ = runCommand("sh", "-c", cleanCmd)
}

// proxyPath returns a PATH that prepends passwall's known binary directories
// (xray / sing-box / lua / app.sh helpers) to whatever the frpc process
// inherited. app.sh run_socks needs these to locate the proxy cores; frpc is
// usually started by init.d with a stripped PATH, so without this the proxy
// fails to start and only logs the error to syslog (silent stdout -> exit 1).
func proxyPath() string {
	base := os.Getenv("PATH")
	extra := "/usr/share/passwall:/usr/share/passwall/usr/bin:/usr/share/passwall/bin:/usr/bin:/bin:/usr/sbin:/sbin"
	if base == "" {
		return extra
	}
	return extra + ":" + base
}

// runProxyCmd runs a shell command with a PATH that includes passwall's tool
// directories, so app.sh run_socks can locate xray/sing-box/lua etc.
func runProxyCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	env := make([]string, 0, len(os.Environ())+1)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "PATH=") {
			env = append(env, e)
		}
	}
	env = append(env, "PATH="+proxyPath())
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// diagnoseProxyStart collects the real failure reason for a failed temp-proxy
// start. passwall logs via echolog (syslog), so we tail logread for passwall
// lines and dump the node's uci config to make the error actionable.
func diagnoseProxyStart(secID string) string {
	var sb strings.Builder
	if out, _ := runCommand("logread"); out != "" {
		var rel []string
		for _, l := range strings.Split(out, "\n") {
			if strings.Contains(l, "passwall") {
				rel = append(rel, l)
			}
		}
		if len(rel) > 40 {
			rel = rel[len(rel)-40:]
		}
		if len(rel) > 0 {
			sb.WriteString("\n--- passwall 日志(tail) ---\n")
			sb.WriteString(strings.Join(rel, "\n"))
		}
	}
	if cfg, _ := runCommand("uci", "-q", "show", "passwall."+secID); cfg != "" {
		sb.WriteString("\n--- 节点配置 ---\n")
		sb.WriteString(cfg)
	}
	return sb.String()
}

// hasRemainingNodes checks if any passwall node sections exist.
func hasRemainingNodes() bool {
	output, err := runCommand("uci", "-q", "show", "passwall")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, "=nodes") {
			secID := strings.TrimSuffix(strings.TrimPrefix(line, "passwall."), "=nodes")
			if secID != "" && secID != "@global[0]" {
				return true
			}
		}
	}
	return false
}

// passwallIsRunning reports whether the passwall service is actually running
// (not just enabled in uci). Used to distinguish “选中了节点但没启动” from “代理中”.
func passwallIsRunning() bool {
	if out, err := runCommand("/etc/init.d/passwall", "status"); err == nil {
		s := strings.ToLower(strings.TrimSpace(out))
		if s != "" && (strings.Contains(s, "running") || strings.Contains(s, "up")) {
			return true
		}
	}
	// Fallback: any passwall-related process present. 注意 passwall 主进程命令行
	// 通常不含 "/var/etc/passwall" 这个路径，故用更宽的 "passwall" 关键字匹配。
	if p, perr := runCommand("pgrep", "-f", "passwall"); perr == nil && strings.TrimSpace(p) != "" {
		return true
	}
	return false
}

// cmdModifyFrp modifies the frp client configuration via uci (OpenWrt).
// The payload is a JSON object with optional fields:
//
//	{"user": "newUser", "serverAddr": "1.2.3.4", "serverPort": 7000}
//
// It modifies /etc/config/frpc (UCI format, section 'common') and then
// restarts frpc in a goroutine. The response is returned BEFORE the restart
// completes, because restarting frpc tears down the control connection
// that would otherwise carry the response back to frps.
func (e *builtinCommandExecutor) cmdModifyFrp(payload string) (string, string) {
	if payload == "" {
		return "error", "错误: 请输入Frp配置"
	}

	var req struct {
		User       string `json:"user"`
		ServerAddr string `json:"serverAddr"`
		ServerPort int    `json:"serverPort"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		// Backwards compat: treat plain string as username only.
		req.User = strings.TrimSpace(payload)
		req.ServerAddr = ""
		req.ServerPort = 0
	}

	var changes []string

	if req.ServerAddr != "" {
		uciSet("frpc.common.server_addr", req.ServerAddr)
		changes = append(changes, fmt.Sprintf("服务地址=%s", req.ServerAddr))
	}
	if req.ServerPort != 0 {
		uciSet("frpc.common.server_port", strconv.Itoa(req.ServerPort))
		changes = append(changes, fmt.Sprintf("服务端口=%d", req.ServerPort))
	}
	if req.User != "" {
		// Reject pure numeric usernames: UCI→TOML conversion would produce
		// `user = 123` (unquoted number), causing frpc to fail with
		// "cannot unmarshal number into string".
		if _, err := strconv.Atoi(req.User); err == nil {
			return "error", "错误: 用户名不能是纯数字"
		}
		uciSet("frpc.common.user", req.User)
		changes = append(changes, fmt.Sprintf("用户名=%s", req.User))
	}

	if len(changes) == 0 {
		return "error", "错误: 未提供任何修改项"
	}

	uciCommit("frpc")

	// Restart frpc in a goroutine: the restart tears down the current
	// control connection, so if we wait for it the response would never
	// reach frps (the channel is gone). Return success now, restart async.
	go func() {
		time.Sleep(200 * time.Millisecond) // give the response time to be sent
		if _, err := runCommand("/etc/init.d/frpc", "restart"); err != nil {
			log.Warnf("frpc restart failed: %v", err)
		}
	}()

	return "ok", fmt.Sprintf("已修改Frp配置: %s，frpc 即将重启", strings.Join(changes, ", "))
}

// cmdModifySystem handles network/wireless tunables on OpenWrt:
//   - toggling the WAN6 interface via network.wan6.disabled
//   - toggling 2.4G / 5G WiFi radio (disabled option)
//   - renaming the WiFi SSID (ssid option)
//
// Payload (all fields optional):
//
//	{
//	  "wan6":   true|false,   // enable/disable the wan6 interface
//	  "wifi2g": true|false,   // enable/disable the 2.4G wifi-iface
//	  "wifi5g": true|false,   // enable/disable the 5G wifi-iface
//	  "ssid":   "MyWiFi"      // rename the SSID on both bands
//	}
//
// uciBoolTrue reports whether the given uci option means "enabled".
// An empty value (option unset), "0" or "off" => enabled (true);
// "1" or "on" => disabled (false).
func uciBoolEnabled(key string) bool {
	v := strings.TrimSpace(uciGet(key))
	if v == "" || v == "0" || v == "off" || v == "false" {
		return true
	}
	return false
}

// wifiBand groups a wifi-iface uci section together with a human label.
type wifiBand struct {
	section string // uci section, e.g. "wireless.@wifi-iface[0]"
	label   string // human readable, e.g. "2.4G", "5G", "5.8G"
	order   int    // sort weight: 2.4G < 5G < 5.8G < 6G
}

// bandOrder returns a sort weight so bands render in a stable, intuitive
// order: 2.4G first, then 5G, then 5.8G, then 6G, unknown bands last.
func bandOrder(label string) int {
	switch label {
	case "2.4G":
		return 0
	case "5G":
		return 1
	case "5.8G":
		return 2
	case "6G":
		return 3
	}
	return 9
}

// baseBandLabel turns a raw uci `band` value into a friendly base label.
func baseBandLabel(band string) string {
	switch strings.ToLower(band) {
	case "2g", "2.4g", "b", "g", "n", "gn", "bgn", "bgnax":
		return "2.4G"
	case "5g", "5", "a", "ac", "ax", "an", "ana", "anac":
		return "5G"
	case "5.8g", "58g", "5.8":
		return "5.8G"
	case "6g", "6":
		return "6G"
	}
	return band
}

// channelToLabel decides whether a 5g radio is the common 5.2G band or the
// 5.8G band. OpenWrt exposes both as band "5g" and only differs by channel:
// low channels (36-64) are 5.2G, high channels (>=149) are 5.8G in most
// regions. A non-numeric / "auto" channel can't be distinguished, so it stays
// "5G".
func channelToLabel(section, dev string) string {
	ch := strings.TrimSpace(uciGet(section + ".channel"))
	if ch == "" && dev != "" {
		ch = strings.TrimSpace(uciGet("wireless." + dev + ".channel"))
		if ch == "" {
			ch = strings.TrimSpace(uciGet(dev + ".channel"))
		}
	}
	n, err := strconv.Atoi(ch)
	if err != nil {
		return "5G"
	}
	if n >= 149 {
		return "5.8G"
	}
	return "5G"
}

// collectWifiIfaces returns every wifi-iface as a wifiBand. Band is detected
// from the iface's own `band` option, otherwise from the referenced
// wifi-device, otherwise by declaration order (2.4G, 5G, 5.8G, ...).
//
// For 5g radios whose band alone can't tell 5.2G from 5.8G (both report
// "5g"), the channel is inspected so a "5.8G" radio is labeled accordingly.
func (e *builtinCommandExecutor) collectWifiIfaces() []wifiBand {
	out, err := runCommand("sh", "-c",
		"uci -q show wireless | sed -n 's/^\\([^=]*\\)=wifi-iface/\\1/p'")
	if err != nil || strings.TrimSpace(out) == "" {
		return nil
	}
	ifaces := strings.Split(strings.TrimSpace(out), "\n")

	bands := make([]wifiBand, 0, len(ifaces))
	for _, s := range ifaces {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		dev := strings.TrimSpace(uciGet(s + ".device"))
		band := strings.TrimSpace(uciGet(s + ".band"))
		if band == "" && dev != "" {
			band = strings.TrimSpace(uciGet("wireless." + dev + ".band"))
			if band == "" {
				band = strings.TrimSpace(uciGet(dev + ".band"))
			}
		}
		label := baseBandLabel(band)
		if label == "5G" {
			// Refine 5.2G vs 5.8G using the channel.
			label = channelToLabel(s, dev)
		}
		if label == "" {
			label = "WiFi"
		}
		bands = append(bands, wifiBand{section: s, label: label, order: bandOrder(label)})
	}

	// Stable order: 2.4G, 5G, 5.8G, 6G, then anything else.
	sort.SliceStable(bands, func(i, j int) bool {
		if bands[i].order != bands[j].order {
			return bands[i].order < bands[j].order
		}
		return bands[i].section < bands[j].section
	})
	return bands
}

// defaultRootShadow is the expected content of the root line in /etc/shadow
// when the device still uses the factory default password.
const defaultRootShadow = "root:$5$MZloauSqpcvpjtZb$NuVJ6qEGPkanc7/986bDfZnF22V43GXfxl00hhremR4:20440:0:99999:7:::"

// shadowPath / shadowBackupPath: OpenWrt stores credentials in /etc/shadow.
// passwd simply rewrites this file, so OpenWrt needs no extra "apply" step —
// writing the file takes effect immediately.
const shadowPath = "/etc/shadow"
const shadowBackupPath = "/etc/shadow.bk"

// defaultPasswordMu guards the 1-minute auto-restore timer.
var defaultPasswordMu sync.Mutex

// defaultPasswordCancel cancels a pending auto-restore (if any).
var defaultPasswordCancel context.CancelFunc

// readShadowRootLine returns the "root:..." line from /etc/shadow (empty if absent).
func readShadowRootLine() string {
	data, err := os.ReadFile(shadowPath)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "root:") {
			return line
		}
	}
	return ""
}

// cmdGetDefaultPassword reports whether the root password is still the default.
//
//	{ "isDefault": bool }
func (e *builtinCommandExecutor) cmdGetDefaultPassword() (string, string) {
	rootLine := readShadowRootLine()
	isDefault := rootLine == defaultRootShadow
	result := map[string]any{"isDefault": isDefault}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "error", fmt.Sprintf("错误: 序列化失败: %v", err)
	}
	return "ok", string(jsonBytes)
}

// cmdSetDefaultPassword enables or disables the default password.
//
// Payload:
//
//	{ "enable": bool }   // true => restore default password; false => restore backup
//
// Behaviour:
//   - On enable (true): if the current root password is NOT the default, back up
//     the current /etc/shadow to /etc/shadow.bk (creating it if missing), then
//     overwrite /etc/shadow with the default password line. A 1-minute timer is
//     started; when it expires the backup is restored automatically.
//   - On disable (false): cancel the pending timer (if any) and immediately
//     restore the backup, so the original password is kept.
func (e *builtinCommandExecutor) cmdSetDefaultPassword(payload string) (string, string) {
	var req struct {
		Enable *bool `json:"enable"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "error", fmt.Sprintf("错误: JSON 解析失败: %v", err)
	}
	if req.Enable == nil {
		return "error", "错误: 缺少 enable 字段"
	}

	defaultPasswordMu.Lock()
	defer defaultPasswordMu.Unlock()

	if !*req.Enable {
		// Disable: cancel the pending timer and restore the backup now.
		if defaultPasswordCancel != nil {
			defaultPasswordCancel()
			defaultPasswordCancel = nil
		}
		if err := restoreShadowBackup(); err != nil {
			return "error", err.Error()
		}
		return "ok", "已取消默认密码并恢复备份"
	}

	// Enable: check current state.
	rootLine := readShadowRootLine()
	if rootLine == defaultRootShadow {
		return "ok", "当前已是默认密码，无需修改"
	}

	// Back up current shadow (create /etc/shadow.bk if missing).
	if err := backupShadow(); err != nil {
		return "error", err.Error()
	}

	// Write the default password line into /etc/shadow.
	if err := writeShadowRootLine(defaultRootShadow); err != nil {
		return "error", fmt.Sprintf("错误: 写入默认密码失败: %v", err)
	}

	// Start a 1-minute auto-restore timer.
	if defaultPasswordCancel != nil {
		defaultPasswordCancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defaultPasswordCancel = cancel
	go func() {
		timer := time.NewTimer(1 * time.Minute)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			// Cancelled (operator disabled early): keep backup as-is.
			return
		case <-timer.C:
			defaultPasswordMu.Lock()
			defaultPasswordCancel = nil
			defaultPasswordMu.Unlock()
			if err := restoreShadowBackup(); err != nil {
				log.Warnf("auto-restore default password failed: %v", err)
				return
			}
			log.Infof("default password auto-restored after 1 minute")
		}
	}()

	return "ok", "已切换为默认密码，1 分钟后将自动恢复"
}

// backupShadow copies /etc/shadow to /etc/shadow.bk (overwriting any existing backup).
func backupShadow() error {
	data, err := os.ReadFile(shadowPath)
	if err != nil {
		return fmt.Errorf("错误: 读取 /etc/shadow 失败: %v", err)
	}
	if err := os.WriteFile(shadowBackupPath, data, 0600); err != nil {
		return fmt.Errorf("错误: 备份 /etc/shadow 失败: %v", err)
	}
	return nil
}

// restoreShadowBackup restores /etc/shadow from /etc/shadow.bk if the backup exists.
func restoreShadowBackup() error {
	if _, err := os.Stat(shadowBackupPath); err != nil {
		// No backup to restore.
		return nil
	}
	data, err := os.ReadFile(shadowBackupPath)
	if err != nil {
		return fmt.Errorf("错误: 读取备份失败: %v", err)
	}
	if err := os.WriteFile(shadowPath, data, 0600); err != nil {
		return fmt.Errorf("错误: 恢复备份失败: %v", err)
	}
	return nil
}

// writeShadowRootLine replaces the root line in /etc/shadow with the given line,
// preserving all other accounts. If no root line exists, it is appended.
func writeShadowRootLine(newRootLine string) error {
	data, err := os.ReadFile(shadowPath)
	if err != nil {
		return fmt.Errorf("错误: 读取 /etc/shadow 失败: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "root:") {
			lines[i] = newRootLine
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, newRootLine)
	}
	out := strings.Join(lines, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	if err := os.WriteFile(shadowPath, []byte(out), 0600); err != nil {
		return fmt.Errorf("错误: 写入 /etc/shadow 失败: %v", err)
	}
	return nil
}

// cmdGetSystem reports the current network/wireless state as JSON:
//
//	{
//	  "wan6": bool,
//	  "bands": [ { "key": "...", "label": "2.4G", "enabled": true, "ssid": "x" } ],
//	  "ssid": string   // first band's ssid (kept for convenience)
//	}
func (e *builtinCommandExecutor) cmdGetSystem() (string, string) {
	wan6 := uciBoolEnabled("network.wan6.disabled")

	bands := e.collectWifiIfaces()
	bandOut := make([]map[string]any, 0, len(bands))
	firstSsid := ""
	for i, b := range bands {
		enabled := uciBoolEnabled(b.section + ".disabled")
		ssid := uciGet(b.section + ".ssid")
		if i == 0 {
			firstSsid = ssid
		}
		bandOut = append(bandOut, map[string]any{
			"key":     b.section,
			"label":   b.label,
			"enabled": enabled,
			"ssid":    ssid,
		})
	}

	result := map[string]any{
		"wan6":  wan6,
		"bands": bandOut,
		"ssid":  firstSsid,
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "error", fmt.Sprintf("错误: 序列化失败: %v", err)
	}
	return "ok", string(jsonBytes)
}

// cmdModifySystem handles network/wireless tunables on OpenWrt:
//   - toggling the WAN6 interface via network.wan6.disabled
//   - toggling 2.4G / 5G WiFi radio (disabled option)
//   - renaming the WiFi SSID (ssid option)
//
// Payload (all fields optional):
//
//	{
//	  "wan6":   true|false,   // enable/disable the wan6 interface
//	  "wifi2g": true|false,   // enable/disable the 2.4G wifi-iface
//	  "wifi5g": true|false,   // enable/disable the 5G wifi-iface
//	  "ssid":   "MyWiFi"      // rename the SSID on both bands
//	}
//
// A field is only applied when present in the JSON.
func (e *builtinCommandExecutor) cmdModifySystem(payload string) (string, string) {
	if strings.TrimSpace(payload) == "" {
		return "error", "错误: 请输入系统设置"
	}

	var req struct {
		WAN6  *bool  `json:"wan6"`
		SSID  string `json:"ssid"`
		Bands []struct {
			Key     string `json:"key"`
			Enabled *bool  `json:"enabled"`
		} `json:"bands"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "error", fmt.Sprintf("错误: 系统设置 JSON 解析失败: %v", err)
	}

	hasChange := req.WAN6 != nil || req.SSID != "" || len(req.Bands) > 0
	if !hasChange {
		return "error", "错误: 未提供任何修改项"
	}

	var changes []string
	needNetworkCommit := false
	needWirelessCommit := false

	// 1) WAN6 toggle: set/clear network.wan6.disabled.
	//    disabled='1' -> off, delete it (or '0') -> on.
	if req.WAN6 != nil {
		state := "关闭"
		if *req.WAN6 {
			// Remove the disabled option -> interface enabled.
			uciDelete("network.wan6.disabled")
			state = "开启"
		} else {
			uciSet("network.wan6.disabled", "1")
		}
		needNetworkCommit = true
		changes = append(changes, fmt.Sprintf("WAN6=%s", state))
	}

	// 2) Per-band WiFi toggle + 3) SSID rename via uci wifi-iface.
	if len(req.Bands) > 0 || req.SSID != "" {
		known := e.collectWifiIfaces()
		sectionByKey := make(map[string]string, len(known))
		labelByKey := make(map[string]string, len(known))
		for _, b := range known {
			sectionByKey[b.section] = b.section
			labelByKey[b.section] = b.label
		}

		for _, band := range req.Bands {
			section, ok := sectionByKey[band.Key]
			if !ok || section == "" {
				changes = append(changes, fmt.Sprintf("%s=未找到", band.Key))
				continue
			}
			label := labelByKey[section]
			if band.Enabled != nil {
				disabled := "0"
				state := "开启"
				if !*band.Enabled {
					disabled = "1"
					state = "关闭"
				}
				uciSet(section+".disabled", disabled)
				changes = append(changes, fmt.Sprintf("%s WiFi=%s", label, state))
			}
			if req.SSID != "" {
				uciSet(section+".ssid", req.SSID)
				changes = append(changes, fmt.Sprintf("%s SSID=%s", label, req.SSID))
			}
		}

		needWirelessCommit = true
	}

	// Commit uci configs and apply changes.
	if needNetworkCommit {
		uciCommit("network")
		// Bring the wan6 interface up so the committed config takes effect.
		if req.WAN6 != nil && *req.WAN6 {
			if _, err := runCommand("if", "up", "wan6"); err != nil {
				log.Warnf("ifup wan6 failed: %v", err)
			}
		}
	}
	if needWirelessCommit {
		uciCommit("wireless")
		// Apply wireless changes asynchronously so the command response
		// is not blocked by the (slow) wifi reload.
		go func() {
			time.Sleep(200 * time.Millisecond)
			if _, err := runCommand("wifi", "reload"); err != nil {
				log.Warnf("wifi reload failed: %v", err)
			}
		}()
	}

	return "ok", fmt.Sprintf("已修改系统设置: %s", strings.Join(changes, ", "))
}

// ============================================================
// Firmware update commands
// ============================================================

// ghProxy is the GitHub download proxy for Chinese clients.
// API calls now go directly to api.github.com from the overseas frps server.
// Only download URLs use this proxy.
const ghProxy = "https://gh.2026178.xyz"

// repoMapping maps OpenWrt target to GitHub API base URL (direct, no proxy).
// The frps server (overseas) calls these URLs directly.
var repoMapping = map[string]string{
	"qualcommax": "https://api.github.com/repos/laosan-xx/OpenWRT-CI-VIKINGYFY",
	"ipq60":      "https://api.github.com/repos/laosan-xx/OpenWRT-CI-VIKINGYFY",
	"mediatek":   "https://api.github.com/repos/laosan-xx/CloseWRT-CI",
	"filogic":    "https://api.github.com/repos/laosan-xx/CloseWRT-CI",
}

// cmdDetectPlatform detects the OpenWrt platform and board info.
func (e *builtinCommandExecutor) cmdDetectPlatform() (string, string) {
	target := readDistribTarget()
	if target == "" {
		return "error", "无法读取 /etc/openwrt_release，请确认运行环境为 OpenWrt"
	}

	boardName, model := getBoardInfo()
	if boardName == "" {
		return "error", "无法通过 ubus 获取 board_name"
	}
	boardModel := strings.ReplaceAll(boardName, ",", "_")

	var repoApi string
	for key, api := range repoMapping {
		if strings.Contains(target, key) {
			repoApi = api
			break
		}
	}
	if repoApi == "" {
		return "error", fmt.Sprintf("未支持的平台: %s", target)
	}

	result := map[string]string{
		"target":     target,
		"boardName":  boardName,
		"model":      model,
		"boardModel": boardModel,
		"repoApi":    repoApi,
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "error", fmt.Sprintf("JSON 序列化失败: %v", err)
	}
	return "ok", string(jsonBytes)
}

// cmdGetSystemVersion reads the LuCI status page fragment
// (/www/luci-static/resources/view/status/include/10_system.js) and extracts the
// firmware build version. The page renders a line like:
//
//	ImmortalWRT SNAPSHOT r0-a4638cd / LuCI Master 26.208.15155~5fa57e3 / laosan-xx-26.07.28-03.02.24
//
// where the last "/" separated segment starting with "laosan-" is the firmware
// version. We match that segment directly.
func (e *builtinCommandExecutor) cmdGetSystemVersion() (string, string) {
	const sysJsPath = "/www/luci-static/resources/view/status/include/10_system.js"
	content, err := os.ReadFile(sysJsPath)
	if err != nil {
		return "error", fmt.Sprintf("读取系统版本文件失败: %v", err)
	}

	// Matches the version segment such as laosan-xx-26.07.28-03.02.24.
	re := regexp.MustCompile(`\blaosan-[A-Za-z0-9_.-]+`)
	matches := re.FindAllString(string(content), -1)
	if len(matches) == 0 {
		return "error", "未在该文件中找到系统版本信息"
	}

	result := map[string]string{
		"version": strings.TrimSpace(matches[len(matches)-1]),
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "error", fmt.Sprintf("JSON 序列化失败: %v", err)
	}
	return "ok", string(jsonBytes)
}

// downloadState tracks firmware download progress.
type downloadState struct {
	Status          string  `json:"status"` // idle, downloading, complete, error, cancelled
	Filename        string  `json:"filename"`
	TotalBytes      int64   `json:"totalBytes"`
	DownloadedBytes int64   `json:"downloadedBytes"`
	Progress        float64 `json:"progress"`
	Error           string  `json:"error,omitempty"`
}

var (
	fwDownloadMu     sync.Mutex
	fwDownloadState  = downloadState{Status: "idle"}
	fwDownloadCancel context.CancelFunc
)

// cmdDownloadFirmware starts an async download of a firmware file.
func (e *builtinCommandExecutor) cmdDownloadFirmware(payload string) (string, string) {
	var req struct {
		URL      string `json:"url"`
		Filename string `json:"filename"`
	}
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return "error", fmt.Sprintf("解析参数失败: %v", err)
	}
	if req.URL == "" || req.Filename == "" {
		return "error", "参数不完整: url 和 filename 均必填"
	}

	fwDownloadMu.Lock()
	if fwDownloadState.Status == "downloading" {
		fwDownloadMu.Unlock()
		return "error", "已有下载任务正在进行"
	}
	ctx, cancel := context.WithCancel(context.Background())
	fwDownloadState = downloadState{
		Status:   "downloading",
		Filename: req.Filename,
	}
	fwDownloadCancel = cancel
	fwDownloadMu.Unlock()

	go doDownload(ctx, req.URL, req.Filename)

	result := map[string]string{"status": "downloading", "filename": req.Filename}
	jsonBytes, _ := json.Marshal(result)
	return "ok", string(jsonBytes)
}

func doDownload(ctx context.Context, rawURL, filename string) {
	defer func() {
		if r := recover(); r != nil {
			fwDownloadMu.Lock()
			fwDownloadState.Status = "error"
			fwDownloadState.Error = fmt.Sprintf("panic: %v", r)
			fwDownloadMu.Unlock()
		}
	}()

	safeName := filepath.Base(filename)
	destPath := "/tmp/" + safeName

	client := &http.Client{Timeout: 0} // no timeout for large downloads
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		fwDownloadMu.Lock()
		fwDownloadState.Status = "error"
		fwDownloadState.Error = fmt.Sprintf("创建请求失败: %v", err)
		fwDownloadMu.Unlock()
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fwDownloadMu.Lock()
		if ctx.Err() != nil {
			fwDownloadState.Status = "cancelled"
		} else {
			fwDownloadState.Status = "error"
			fwDownloadState.Error = fmt.Sprintf("下载请求失败: %v", err)
		}
		fwDownloadMu.Unlock()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fwDownloadMu.Lock()
		fwDownloadState.Status = "error"
		fwDownloadState.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		fwDownloadMu.Unlock()
		return
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		fwDownloadMu.Lock()
		fwDownloadState.Status = "error"
		fwDownloadState.Error = fmt.Sprintf("创建文件失败: %v", err)
		fwDownloadMu.Unlock()
		return
	}
	defer outFile.Close()

	total := resp.ContentLength
	fwDownloadMu.Lock()
	fwDownloadState.TotalBytes = total
	fwDownloadMu.Unlock()

	buf := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := outFile.Write(buf[:n]); writeErr != nil {
				fwDownloadMu.Lock()
				fwDownloadState.Status = "error"
				fwDownloadState.Error = fmt.Sprintf("写入文件失败: %v", writeErr)
				fwDownloadMu.Unlock()
				return
			}
			downloaded += int64(n)
			var progress float64
			if total > 0 {
				progress = float64(downloaded) / float64(total) * 100
			}
			fwDownloadMu.Lock()
			fwDownloadState.DownloadedBytes = downloaded
			fwDownloadState.Progress = progress
			fwDownloadMu.Unlock()
		}
		if readErr != nil {
			if readErr != io.EOF {
				fwDownloadMu.Lock()
				// Check if cancelled
				if ctx.Err() != nil {
					fwDownloadState.Status = "cancelled"
				} else {
					fwDownloadState.Status = "error"
					fwDownloadState.Error = fmt.Sprintf("读取响应失败: %v", readErr)
				}
				fwDownloadMu.Unlock()
				// Clean up partial file
				os.Remove(destPath)
				return
			}
			break
		}
		// Check cancellation between reads
		if ctx.Err() != nil {
			fwDownloadMu.Lock()
			fwDownloadState.Status = "cancelled"
			fwDownloadMu.Unlock()
			os.Remove(destPath)
			return
		}
	}

	fwDownloadMu.Lock()
	fwDownloadState.Status = "complete"
	fwDownloadState.Progress = 100
	fwDownloadState.DownloadedBytes = downloaded
	fwDownloadMu.Unlock()
}

// cmdCancelDownload cancels an in-progress firmware download.
func (e *builtinCommandExecutor) cmdCancelDownload() (string, string) {
	fwDownloadMu.Lock()
	if fwDownloadState.Status == "downloading" && fwDownloadCancel != nil {
		fwDownloadCancel()
		fwDownloadCancel = nil
	}
	fwDownloadMu.Unlock()
	return "ok", `{"status":"cancelled"}`
}

// cmdDownloadStatus returns current download progress.
func (e *builtinCommandExecutor) cmdDownloadStatus() (string, string) {
	fwDownloadMu.Lock()
	state := fwDownloadState
	fwDownloadMu.Unlock()

	jsonBytes, _ := json.Marshal(state)
	return "ok", string(jsonBytes)
}

// cmdRunSysupgrade executes sysupgrade on the downloaded firmware.
func (e *builtinCommandExecutor) cmdRunSysupgrade(payload string) (string, string) {
	if payload == "" {
		return "error", "错误: 请指定固件文件名"
	}
	safeName := filepath.Base(payload)
	fwPath := "/tmp/" + safeName

	if _, err := os.Stat(fwPath); err != nil {
		return "error", fmt.Sprintf("固件文件不存在: %s", fwPath)
	}

	log.Infof("sysupgrade: starting upgrade with %s", fwPath)
	// sysupgrade will reboot the device, so this may not return
	cmd := exec.Command("sysupgrade", "-n", fwPath)
	cmd.Start()

	result := map[string]string{"status": "upgrading", "message": "系统更新中，路由器即将重启..."}
	jsonBytes, _ := json.Marshal(result)
	return "ok", string(jsonBytes)
}

// ============================================================
// Platform detection helpers
// ============================================================

func readDistribTarget() string {
	f, err := os.Open("/etc/openwrt_release")
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "DISTRIB_TARGET=") {
			val := strings.TrimPrefix(line, "DISTRIB_TARGET=")
			return strings.Trim(val, "'\"")
		}
	}
	return ""
}

func getBoardInfo() (boardName, model string) {
	out, err := runCommand("ubus", "call", "system", "board")
	if err != nil {
		return "", ""
	}
	// Parse JSON output from ubus
	var board struct {
		BoardName string `json:"board_name"`
		Model     string `json:"model"`
	}
	if err := json.Unmarshal([]byte(out), &board); err != nil {
		// Fallback: grep-style extraction
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "\"board_name\"") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					boardName = strings.Trim(strings.TrimSpace(parts[1]), "\",\"")
				}
			}
			if strings.Contains(line, "\"model\"") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					model = strings.Trim(strings.TrimSpace(parts[1]), "\",\"")
				}
			}
		}
		return boardName, model
	}
	return board.BoardName, board.Model
}

// ============================================================
// Share link parsing
// ============================================================

type parsedNode struct {
	Remarks  string
	Type     string // passwall backend type: "Xray" or "sing-box"
	Protocol string // protocol: vmess, vless, shadowsocks, trojan
	Address  string
	Port     string
	Extra    map[string]string // additional uci fields
}

func parseShareLink(link string) (*parsedNode, error) {
	switch {
	case strings.HasPrefix(link, "ss://"):
		return parseSSLink(link)
	case strings.HasPrefix(link, "vmess://"):
		return parseVmessLink(link)
	case strings.HasPrefix(link, "vless://"):
		return parseVlessLink(link)
	case strings.HasPrefix(link, "trojan://"):
		return parseTrojanLink(link)
	default:
		return nil, fmt.Errorf("不支持的链接格式: %s", link)
	}
}

// parseSSLink parses ss:// links (both legacy base64 and SIP002 format).
func parseSSLink(link string) (*parsedNode, error) {
	body := strings.TrimPrefix(link, "ss://")

	// Try SIP002 format first: ss://BASE64(method:password)@host:port#tag
	if strings.Contains(body, "@") {
		u, err := url.Parse("ss://" + body)
		if err != nil {
			return nil, err
		}
		// userinfo is base64(method:password)
		userInfo := ""
		if u.User != nil {
			userInfo = u.User.Username()
			if p, ok := u.User.Password(); ok {
				userInfo += ":" + p
			}
		}
		decoded, err := base64.RawURLEncoding.DecodeString(userInfo)
		if err != nil {
			// Try standard base64
			decoded, err = base64.StdEncoding.DecodeString(userInfo)
			if err != nil {
				return nil, fmt.Errorf("解码SS用户信息失败: %v", err)
			}
		}
		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("SS链接格式错误")
		}
		method, password := parts[0], parts[1]
		remarks := u.Fragment
		if remarks == "" {
			remarks = u.Host
		}
		remarks, _ = url.QueryUnescape(remarks)

		extra := map[string]string{
			"method":    method,
			"password":  password,
			"transport": "raw",
		}
		// Parse simple-obfs plugin (SIP002) if present.
		if plugin := u.Query().Get("plugin"); plugin != "" {
			// plugin looks like "simple-obfs;obfs=http;obfs-host=xxx;obfs-param=yyy"
			if strings.HasPrefix(plugin, "simple-obfs") {
				extra["plugin"] = "obfs"
				for _, p := range strings.Split(plugin, ";")[1:] {
					kv := strings.SplitN(p, "=", 2)
					if len(kv) != 2 {
						continue
					}
					switch kv[0] {
					case "obfs":
						extra["obfs"] = kv[1]
					case "obfs-host":
						extra["obfs_host"] = kv[1]
					case "obfs-param":
						extra["obfs_param"] = kv[1]
					}
				}
			} else {
				extra["plugin"] = plugin
			}
		}

		return &parsedNode{
			Remarks:  remarks,
			Type:     "Xray",
			Protocol: "shadowsocks",
			Address:  u.Hostname(),
			Port:     u.Port(),
			Extra:    extra,
		}, nil
	}

	// Legacy format: ss://BASE64(method:password@host:port)
	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(body)
		if err != nil {
			return nil, fmt.Errorf("解码SS链接失败: %v", err)
		}
	}
	// format: method:password@host:port
	atIdx := strings.LastIndex(string(decoded), "@")
	if atIdx < 0 {
		return nil, fmt.Errorf("SS链接格式错误: 缺少@")
	}
	userInfo := string(decoded[:atIdx])
	hostPort := string(decoded[atIdx+1:])

	parts := strings.SplitN(userInfo, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("SS链接格式错误")
	}
	method, password := parts[0], parts[1]

	host, port, err := splitHostPort(hostPort)
	if err != nil {
		return nil, err
	}

	return &parsedNode{
		Remarks:  host,
		Type:     "Xray",
		Protocol: "shadowsocks",
		Address:  host,
		Port:     port,
		Extra: map[string]string{
			"method":    method,
			"password":  password,
			"transport": "raw",
		},
	}, nil
}

// parseVmessLink parses vmess:// links (base64 JSON).
func parseVmessLink(link string) (*parsedNode, error) {
	body := strings.TrimPrefix(link, "vmess://")
	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(body)
		if err != nil {
			return nil, fmt.Errorf("解码vmess链接失败: %v", err)
		}
	}

	var vm map[string]any
	if err := json.Unmarshal(decoded, &vm); err != nil {
		return nil, fmt.Errorf("解析vmess JSON失败: %v", err)
	}

	remarks, _ := vm["ps"].(string)
	addr, _ := vm["add"].(string)
	port, _ := vm["port"].(string)
	id, _ := vm["id"].(string)
	aid, _ := vm["aid"].(string)
	net, _ := vm["net"].(string)
	tls, _ := vm["tls"].(string)
	host, _ := vm["host"].(string)
	path, _ := vm["path"].(string)

	if remarks == "" {
		remarks = addr
	}
	if net == "" {
		net = "tcp"
	}
	// Xray uses "raw" instead of "tcp"
	if net == "tcp" {
		net = "raw"
	}

	extra := map[string]string{
		"uuid":      id,
		"transport": net,
	}
	if aid != "" && aid != "0" {
		extra["alter_id"] = aid
	}
	if tls == "tls" {
		extra["tls"] = "1"
	}
	if host != "" {
		extra["ws_host"] = host
	}
	if path != "" {
		extra["ws_path"] = path
	}

	return &parsedNode{
		Remarks:  remarks,
		Type:     "Xray",
		Protocol: "vmess",
		Address:  addr,
		Port:     port,
		Extra:    extra,
	}, nil
}

// parseVlessLink parses vless:// links.
func parseVlessLink(link string) (*parsedNode, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("解析vless链接失败: %v", err)
	}

	uuid := ""
	if u.User != nil {
		uuid = u.User.Username()
	}
	remarks := u.Fragment
	if remarks == "" {
		remarks = u.Hostname()
	}
	remarks, _ = url.QueryUnescape(remarks)

	q := u.Query()
	net := q.Get("type")
	if net == "" {
		net = "tcp"
	}
	// Xray uses "raw" instead of "tcp"
	if net == "tcp" {
		net = "raw"
	}
	security := q.Get("security")

	extra := map[string]string{
		"uuid":      uuid,
		"transport": net,
	}
	if security == "tls" || security == "reality" {
		extra["tls"] = "1"
	}
	if flow := q.Get("flow"); flow != "" {
		extra["flow"] = flow
	}
	// vless encryption (e.g. "none")
	if enc := q.Get("encryption"); enc != "" && enc != "none" {
		extra["encryption"] = enc
	}
	if alpn := q.Get("alpn"); alpn != "" {
		extra["alpn"] = alpn
	}
	sni := q.Get("sni")
	if h := q.Get("host"); h != "" {
		extra["ws_host"] = h
	} else if net == "ws" && sni != "" {
		// For ws transport, fall back to sni as host if no explicit host
		extra["ws_host"] = sni
	}
	if p := q.Get("path"); p != "" {
		extra["ws_path"] = p
	}
	if sni != "" {
		extra["tls_serverName"] = sni
	}
	if fp := q.Get("fp"); fp != "" {
		extra["utls"] = "1"
		extra["fingerprint"] = fp
	}
	if pbk := q.Get("pbk"); pbk != "" {
		extra["reality_publicKey"] = pbk
	}
	if sid := q.Get("sid"); sid != "" {
		extra["reality_shortId"] = sid
	}

	return &parsedNode{
		Remarks:  remarks,
		Type:     "Xray",
		Protocol: "vless",
		Address:  u.Hostname(),
		Port:     u.Port(),
		Extra:    extra,
	}, nil
}

// parseTrojanLink parses trojan:// links.
func parseTrojanLink(link string) (*parsedNode, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("解析trojan链接失败: %v", err)
	}

	password := ""
	if u.User != nil {
		password = u.User.Username()
	}
	remarks := u.Fragment
	if remarks == "" {
		remarks = u.Hostname()
	}
	remarks, _ = url.QueryUnescape(remarks)

	q := u.Query()
	extra := map[string]string{
		"password": password,
	}
	if q.Get("security") == "tls" || q.Get("type") == "" || q.Get("type") == "tcp" {
		extra["tls"] = "1"
	}
	if sni := q.Get("sni"); sni != "" {
		extra["tls_serverName"] = sni
	}

	return &parsedNode{
		Remarks:  remarks,
		Type:     "Xray",
		Protocol: "trojan",
		Address:  u.Hostname(),
		Port:     u.Port(),
		Extra:    extra,
	}, nil
}

// ============================================================
// uci helper functions
// ============================================================

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func uciGet(key string) string {
	out, err := runCommand("uci", "-q", "get", key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func uciSet(key, value string) {
	_, _ = runCommand("uci", "-q", "set", key+"="+value)
}

func uciDelete(key string) {
	_, _ = runCommand("uci", "-q", "delete", key)
}

func uciCommit(config string) {
	_, _ = runCommand("uci", "commit", config)
}

// uciAddSection adds a new named section and returns its name.
func uciAddSection(config, secType string) (string, error) {
	out, err := runCommand("uci", "add", config, secType)
	if err != nil {
		return "", fmt.Errorf("uci add failed: %v, output: %s", err, out)
	}
	name := strings.TrimSpace(out)
	if name == "" {
		return "", fmt.Errorf("uci add 未返回section名")
	}
	return name, nil
}

// findNodeByRemarks finds a passwall node section ID by its remarks.
// Parses everything from a single `uci show` output to avoid format mismatch issues.
func findNodeByRemarks(remarks string) string {
	output, err := runCommand("uci", "-q", "show", "passwall")
	if err != nil {
		return ""
	}

	// Pass 1: build secID -> section type map from "passwall.XXX=TYPE" lines
	secTypes := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "passwall.") {
			continue
		}
		// Match "passwall.XXX=TYPE" (type declaration line, no dots after secID)
		rest := strings.TrimPrefix(line, "passwall.")
		eqIdx := strings.Index(rest, "=")
		if eqIdx < 0 {
			continue
		}
		secID := rest[:eqIdx]
		val := rest[eqIdx+1:]
		// Type declaration lines have no dots in secID (e.g. cfg1a2b3c=nodes)
		// Option lines have dots (e.g. cfg1a2b3c.remarks='xxx')
		if !strings.Contains(secID, ".") {
			secTypes[secID] = val
		}
	}

	// Pass 2: find remarks match and verify type from our map
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, ".remarks=") {
			continue
		}
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}
		key := line[:eqIdx]
		val := strings.Trim(line[eqIdx+1:], "'")
		if val == remarks {
			// key format: "passwall.<secID>.remarks"
			keyParts := strings.SplitN(key, ".", 3)
			if len(keyParts) >= 3 {
				secID := keyParts[1] // secID is the second part
				// Check type from our parsed map
				if secTypes[secID] == "nodes" {
					return secID
				}
				// Handle @nodes[X] format (type is self-evident)
				if strings.HasPrefix(secID, "@nodes[") {
					return secID
				}
			}
		}
	}
	return ""
}

// readSection reads all options of a uci section into a map (key without
// the "passwall.<secID>." prefix). Unquoted values are trimmed.
func readSection(secID string) map[string]string {
	output, err := runCommand("uci", "-q", "show", "passwall."+secID)
	if err != nil {
		return map[string]string{}
	}
	opts := map[string]string{}
	prefix := "passwall." + secID + "."
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimPrefix(line, prefix)
		eqIdx := strings.Index(rest, "=")
		if eqIdx < 0 {
			continue
		}
		key := rest[:eqIdx]
		val := strings.Trim(rest[eqIdx+1:], "'")
		opts[key] = val
	}
	return opts
}

// nodeExport returns the share link of the node identified by remarks.
// It always rebuilds the link live from the current uci options (using the
// same field names parseShareLink writes on import) to avoid serving any
// stale/corrupted cached value.
func nodeExport(remarks string) string {
	secID := findNodeByRemarks(remarks)
	if secID == "" {
		return ""
	}
	return buildShareLink(readSection(secID))
}

// b64enc standard base64 encode (passwall uses StdEncoding for ss/vmess).
func b64enc(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// buildShareLink reconstructs a passwall node share link from its uci options.
// The option keys below strictly mirror what parseShareLink writes into the
// uci Extra map (e.g. "tls_serverName", "reality_publicKey", etc.) so the
// exported link is faithful to the original import.
func buildShareLink(opts map[string]string) string {
	protocol := opts["protocol"]
	remarks := opts["remarks"]

	switch strings.ToLower(protocol) {
	case "ss", "shadowsocks":
		return buildSSLink(opts, remarks)
	case "vmess":
		return buildVMessLink(opts, remarks)
	case "vless":
		return buildVLESSLink(opts, remarks)
	case "trojan":
		return buildTrojanLink(opts, remarks)
	default:
		return ""
	}
}

// transportAndHostPath resolves transport type plus host/path for ws/h2/grpc.
func transportAndHostPath(opts map[string]string) (net, host, path string) {
	net = optOr(opts, "transport", "tcp")
	// parseShareLink stores tcp as "raw" (Xray internal); the standard share
	// link format uses "tcp".
	if net == "raw" {
		net = "tcp"
	}
	switch net {
	case "ws":
		host = optOr(opts, "ws_host", "")
		path = optOr(opts, "ws_path", "")
	case "h2":
		host = optOr(opts, "h2_host", "")
		path = optOr(opts, "h2_path", "")
	case "grpc":
		// passwall stores grpc service name without "grpc_" prefix in some
		// versions; try both.
		path = optOr(opts, "grpc_service_name", optOr(opts, "serviceName", ""))
	}
	return
}

// securityInfo determines whether the node uses reality/tls/none and returns
// the security value plus whether it is reality.
// Reality is identified strictly by the presence of reality-specific fields.
// Note: "utls" only means the utls library is used (common for normal tls
// nodes too), so it must NOT be treated as a reality marker.
func securityInfo(opts map[string]string) (security string, isReality bool) {
	isReality = optOr(opts, "reality_publicKey", "") != "" ||
		optOr(opts, "reality_shortId", "") != ""
	if isReality {
		return "reality", true
	}
	if tlsValue(opts) == "1" {
		return "tls", false
	}
	return "none", false
}

func buildSSLink(opts map[string]string, remarks string) string {
	// cmdNodeLink stores the address under "address" (not "server").
	server := optOr(opts, "address", opts["server"])
	port := optOr(opts, "port", "")
	method := optOr(opts, "method", "")
	password := optOr(opts, "password", "")
	if server == "" || port == "" || method == "" {
		return ""
	}

	// Passwall stores obfs plugin info in "plugin" + "obfs"/"obfs_host"/"obfs_param".
	var pluginOpts string
	if plugin := optOr(opts, "plugin", ""); plugin != "" {
		switch plugin {
		case "obfs":
			obfs := optOr(opts, "obfs", "")
			if obfs != "" {
				pluginOpts = "obfs=" + obfs
				if h := optOr(opts, "obfs_host", ""); h != "" {
					pluginOpts += ";obfs-host=" + h
				}
				if p := optOr(opts, "obfs_param", ""); p != "" {
					pluginOpts += ";obfs-param=" + p
				}
			}
		default:
			pluginOpts = optOr(opts, "plugin_opts", "")
		}
	}

	userInfo := method + ":" + password
	if pluginOpts != "" {
		link := "ss://" + b64enc(userInfo) + "@" + server + ":" + port
		link += "?plugin=" + url.QueryEscape("simple-obfs;"+pluginOpts)
		link += "#" + url.QueryEscape(remarks)
		return link
	}
	link := "ss://" + b64enc(userInfo+"@"+server+":"+port)
	link += "#" + url.QueryEscape(remarks)
	return link
}

func buildVMessLink(opts map[string]string, remarks string) string {
	add := optOr(opts, "address", "")
	port := optOr(opts, "port", "")
	id := optOr(opts, "uuid", "")
	if add == "" || port == "" || id == "" {
		return ""
	}
	aid := optOr(opts, "alter_id", "0")
	scy := optOr(opts, "security", "auto")
	net, host, path := transportAndHostPath(opts)
	typeField := optOr(opts, "tcp_guise", optOr(opts, "header_type", "none"))
	tls := tlsValue(opts)
	sni := optOr(opts, "tls_serverName", optOr(opts, "tls_host", ""))
	allowInsecure := optOr(opts, "tls_insecure", "0")
	if allowInsecure == "1" {
		allowInsecure = "1"
	}

	conf := map[string]interface{}{
		"v":    "2",
		"ps":   remarks,
		"add":  add,
		"port": port,
		"id":   id,
		"aid":  aid,
		"scy":  scy,
		"net":  net,
		"type": typeField,
		"host": host,
		"path": path,
		"tls":  tls,
		"sni":  sni,
		"alpn": optOr(opts, "alpn", ""),
	}
	if allowInsecure == "1" {
		conf["allowInsecure"] = "1"
	}
	if net == "grpc" && path != "" {
		conf["serviceName"] = path
	}
	raw, err := json.Marshal(conf)
	if err != nil {
		return ""
	}
	return "vmess://" + b64enc(string(raw))
}

func buildVLESSLink(opts map[string]string, remarks string) string {
	add := optOr(opts, "address", "")
	port := optOr(opts, "port", "")
	id := optOr(opts, "uuid", "")
	if add == "" || port == "" || id == "" {
		return ""
	}
	net, host, path := transportAndHostPath(opts)
	flow := optOr(opts, "flow", "")
	encryption := optOr(opts, "encryption", "none")
	sni := optOr(opts, "tls_serverName", optOr(opts, "tls_host", ""))
	allowInsecure := optOr(opts, "tls_insecure", "0")
	fp := optOr(opts, "fingerprint", "")
	pbk := optOr(opts, "reality_publicKey", "")
	sid := optOr(opts, "reality_shortId", "")
	alpn := optOr(opts, "alpn", "")

	security, isReality := securityInfo(opts)

	query := map[string]string{}
	query["encryption"] = encryption
	if flow != "" {
		query["flow"] = flow
	}
	query["security"] = security
	query["type"] = net
	if host != "" {
		query["host"] = host
	}
	if path != "" {
		query["path"] = path
	}
	if security != "none" && sni != "" {
		query["sni"] = sni
	}
	if alpn != "" {
		query["alpn"] = alpn
	}
	if allowInsecure == "1" {
		query["allowInsecure"] = "1"
	}
	if isReality {
		if fp != "" {
			query["fp"] = fp
		}
		if pbk != "" {
			query["pbk"] = pbk
		}
		if sid != "" {
			query["sid"] = sid
		}
	}

	return "vless://" + id + "@" + add + ":" + port +
		"?" + buildQuery(query) + "#" + url.QueryEscape(remarks)
}

func buildTrojanLink(opts map[string]string, remarks string) string {
	add := optOr(opts, "address", "")
	port := optOr(opts, "port", "")
	password := optOr(opts, "password", "")
	if add == "" || port == "" || password == "" {
		return ""
	}
	net, host, path := transportAndHostPath(opts)
	sni := optOr(opts, "tls_serverName", optOr(opts, "tls_host", ""))
	allowInsecure := optOr(opts, "tls_insecure", "0")

	query := map[string]string{}
	query["type"] = net
	// trojan always uses tls
	query["security"] = "tls"
	if sni != "" {
		query["sni"] = sni
	}
	if allowInsecure == "1" {
		query["allowInsecure"] = "1"
	}
	if host != "" {
		query["host"] = host
	}
	if path != "" {
		query["path"] = path
	}

	return "trojan://" + url.QueryEscape(password) + "@" + add + ":" + port +
		"?" + buildQuery(query) + "#" + url.QueryEscape(remarks)
}

// optOr returns the option value or fallback.
func optOr(opts map[string]string, key, fallback string) string {
	if v, ok := opts[key]; ok && v != "" {
		return v
	}
	return fallback
}

// tlsValue normalizes the passwall "tls" option into "1"/"0".
func tlsValue(opts map[string]string) string {
	v := optOr(opts, "tls", "0")
	if v == "1" || v == "tls" || strings.EqualFold(v, "true") {
		return "1"
	}
	return "0"
}

// buildQuery builds a URL query string (sorted keys for determinism).
func buildQuery(m map[string]string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+url.QueryEscape(m[k]))
	}
	return strings.Join(parts, "&")
}

func splitHostPort(hostPort string) (string, string, error) {
	idx := strings.LastIndex(hostPort, ":")
	if idx < 0 {
		return "", "", fmt.Errorf("无效的地址端口: %s", hostPort)
	}
	host := hostPort[:idx]
	port := hostPort[idx+1:]
	if _, err := strconv.Atoi(port); err != nil {
		return "", "", fmt.Errorf("无效端口: %s", port)
	}
	return host, port, nil
}
