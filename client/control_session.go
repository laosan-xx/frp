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

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/laosan-xx/frp/pkg/auth"
	v1 "github.com/laosan-xx/frp/pkg/config/v1"
	"github.com/laosan-xx/frp/pkg/msg"
	"github.com/laosan-xx/frp/pkg/proto/wire"
	netpkg "github.com/laosan-xx/frp/pkg/util/net"
	"github.com/laosan-xx/frp/pkg/util/version"
	"github.com/laosan-xx/frp/pkg/vnet"
)

type controlSessionDialer struct {
	ctx context.Context

	common         *v1.ClientCommonConfig
	auth           *auth.ClientAuth
	clientSpec     *msg.ClientSpec
	vnetController *vnet.Controller

	connectorCreator func(context.Context, *v1.ClientCommonConfig) Connector
}

func (d *controlSessionDialer) Dial(previousRunID string) (*SessionContext, error) {
	connector := d.connectorCreator(d.ctx, d.common)
	if err := connector.Open(); err != nil {
		return nil, err
	}

	success := false
	defer func() {
		if !success {
			_ = connector.Close()
		}
	}()

	conn, err := connector.Connect()
	if err != nil {
		return nil, err
	}
	defer func() {
		if !success {
			_ = conn.Close()
		}
	}()

	loginMsg, err := d.buildLoginMsg(previousRunID)
	if err != nil {
		return nil, err
	}

	loginResult, err := d.exchangeLogin(conn, loginMsg)
	if err != nil {
		return nil, err
	}
	loginRespMsg := loginResult.resp
	if loginRespMsg.Error != "" {
		return nil, errors.New(loginRespMsg.Error)
	}

	var controlRW io.ReadWriter = conn
	if d.clientSpec == nil || d.clientSpec.Type != "ssh-tunnel" {
		controlRW, err = d.newControlReadWriter(conn, loginResult.crypto)
		if err != nil {
			return nil, fmt.Errorf("create control crypto read writer: %w", err)
		}
	}

	success = true
	return &SessionContext{
		Common:         d.common,
		RunID:          loginRespMsg.RunID,
		Conn:           msg.NewConn(conn, msg.NewReadWriter(controlRW, d.common.Transport.WireProtocol)),
		Auth:           d.auth,
		Connector:      newMessageConnector(connector, d.common.Transport.WireProtocol),
		VnetController: d.vnetController,
		UDPPacketCodec: loginResult.udpPacketCodec,
	}, nil
}

func (d *controlSessionDialer) buildLoginMsg(previousRunID string) (*msg.Login, error) {
	hostname, _ := os.Hostname()
	loginMsg := &msg.Login{
		Arch:      runtime.GOARCH,
		Os:        runtime.GOOS,
		Hostname:  hostname,
		PoolCount: d.common.Transport.PoolCount,
		User:      d.common.User,
		ClientID:  d.common.ClientID,
		Version:   version.Full(),
		Timestamp: time.Now().Unix(),
		RunID:     previousRunID,
		Metas:     d.common.Metadatas,
	}
	if d.clientSpec != nil {
		loginMsg.ClientSpec = *d.clientSpec
	}

	if err := d.auth.Setter.SetLogin(loginMsg); err != nil {
		return nil, err
	}

	// Report the client's own outbound address. When frps is behind a reverse
	// proxy (e.g. nginx), the server only sees the proxy address (127.0.0.1),
	// so the client sends its real outbound IP for the server to compare against.
	if clientAddr := publicIP(); clientAddr != "" {
		loginMsg.ClientAddr = clientAddr
	}
	return loginMsg, nil
}

// publicIP returns the client's public IP address (the IP assigned by the ISP)
// by querying external services. This is useful when frps is behind a reverse
// proxy and cannot see the client's real public IP.
//
// To avoid passwall / transparent proxy hijacking, bypass-friendly HTTP services
// are tried first. These domains are typically configured as direct-routed in
// passwall's default bypass list, so the query goes out through the real ISP
// interface even when the proxy is on. If none respond (e.g. different bypass
// list config), it falls back to standard HTTPS services which may return the
// proxy egress IP when iptables redirection is active.
func publicIP() string {
	// Bypass-friendly HTTP services (plain-text, typically in passwall direct
	// bypass, fast). Each parser extracts the first IPv4 found in the body.
	type ipService struct {
		url    string
		parser func(string) string // nil → use whole body after trimming
	}
	bypassServices := []ipService{
		{"http://myip.ipip.net", extractFirstIP},
		{"http://cip.cc", extractCipCCIP},
	}
	fallbackServices := []string{
		"https://api.ipify.org",
		"https://ifconfig.me",
		"https://ipinfo.io/ip",
	}
	// Explicitly disable proxy to get the real public IP of this machine,
	// not the proxy server's IP.
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			Proxy: nil,
		},
	}
	for _, svc := range bypassServices {
		ip := tryGetIP(client, svc.url, svc.parser)
		if ip != "" {
			return ip
		}
	}
	for _, url := range fallbackServices {
		ip := tryGetIP(client, url, nil)
		if ip != "" {
			return ip
		}
	}
	return ""
}

// tryGetIP fetches url with client, applies parse to the response body
// (or trims the body directly when parse is nil), and returns a valid IP string
// or empty.
func tryGetIP(client *http.Client, url string, parse func(string) string) string {
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(body))
	if parse != nil {
		text = parse(text)
	}
	if text != "" && net.ParseIP(text) != nil {
		return text
	}
	return ""
}

// extractFirstIP returns the first valid IPv4 found in a mixed-content response
// like "当前 IP：1.2.3.4  来自于：...". If no IP is found returns "".
func extractFirstIP(text string) string {
	for _, field := range strings.Fields(strings.NewReplacer("：", ": ", "：", ": ").Replace(text)) {
		field = strings.TrimRight(field, ",，;:：")
		if ip := strings.TrimSpace(field); net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}

// extractCipCCIP extracts the IP from cip.cc responses like:
//
//	IP      : 1.2.3.4
//	...
func extractCipCCIP(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "IP") {
			continue
		}
		line = strings.TrimPrefix(line, "IP")
		line = strings.TrimLeft(line, " \t:：*")
		line = strings.TrimSpace(line)
		if ip, _, _ := strings.Cut(line, " "); net.ParseIP(ip) != nil {
			return ip
		}
	}
	return ""
}

type loginExchangeResult struct {
	resp           *msg.LoginResp
	crypto         *wire.CryptoContext
	udpPacketCodec string
}

func (d *controlSessionDialer) exchangeLogin(conn net.Conn, loginMsg *msg.Login) (*loginExchangeResult, error) {
	rw := msg.NewV1ReadWriter(conn)
	var wireConn *wire.Conn
	var clientHello wire.ClientHello
	var clientHelloPayload []byte

	if d.common.Transport.WireProtocol == wire.ProtocolV2 {
		if err := wire.WriteMagic(conn); err != nil {
			return nil, err
		}

		wireConn = wire.NewConn(conn)
		rw = msg.NewV2ReadWriterWithConn(wireConn)
		var err error
		clientHello, err = wire.NewClientHello(wire.BootstrapInfo{
			Transport: d.common.Transport.Protocol,
			TLS:       lo.FromPtr(d.common.Transport.TLS.Enable) || d.common.Transport.Protocol == "wss" || d.common.Transport.Protocol == "quic",
			TCPMux:    lo.FromPtr(d.common.Transport.TCPMux),
		})
		if err != nil {
			return nil, err
		}
		clientHelloFrame, err := wire.NewJSONFrame(wire.FrameTypeClientHello, clientHello)
		if err != nil {
			return nil, err
		}
		if err := wireConn.WriteFrame(clientHelloFrame); err != nil {
			return nil, err
		}
		clientHelloPayload = clientHelloFrame.Payload
	}
	if err := rw.WriteMsg(loginMsg); err != nil {
		return nil, err
	}

	_ = conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	defer func() {
		_ = conn.SetReadDeadline(time.Time{})
	}()

	var cryptoContext *wire.CryptoContext
	var udpPacketCodec string
	if wireConn != nil {
		serverHelloFrame, err := wireConn.ReadFrame()
		if err != nil {
			return nil, err
		}
		if serverHelloFrame.Type != wire.FrameTypeServerHello {
			return nil, fmt.Errorf("unexpected frame type %d, want %d", serverHelloFrame.Type, wire.FrameTypeServerHello)
		}
		var serverHello wire.ServerHello
		if err := wireConn.UnmarshalFrame(serverHelloFrame, &serverHello); err != nil {
			return nil, err
		}
		if serverHello.Error != "" {
			return nil, errors.New(serverHello.Error)
		}
		cryptoContext, err = wire.NewClientCryptoContext(clientHelloPayload, serverHelloFrame.Payload)
		if err != nil {
			return nil, err
		}
		udpPacketCodec = serverHello.Selected.Message.UDPPacketCodec
	}

	var loginRespMsg msg.LoginResp
	if err := rw.ReadMsgInto(&loginRespMsg); err != nil {
		return nil, err
	}
	return &loginExchangeResult{
		resp:           &loginRespMsg,
		crypto:         cryptoContext,
		udpPacketCodec: udpPacketCodec,
	}, nil
}

func (d *controlSessionDialer) newControlReadWriter(conn net.Conn, cryptoContext *wire.CryptoContext) (io.ReadWriter, error) {
	if d.common.Transport.WireProtocol == wire.ProtocolV2 {
		if cryptoContext == nil {
			return nil, errors.New("missing v2 crypto negotiation")
		}
		return netpkg.NewAEADCryptoReadWriter(
			conn,
			d.auth.EncryptionKey(),
			netpkg.AEADCryptoRoleClient,
			cryptoContext.Algorithm,
			cryptoContext.TranscriptHash,
		)
	}
	return netpkg.NewCryptoReadWriter(conn, d.auth.EncryptionKey())
}
