package net

import (
	"errors"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

var ErrWebsocketListenerClosed = errors.New("websocket listener closed")

const (
	FrpWebsocketPath = "/~!frp"
)

// wsRealAddrConn wraps a net.Conn to override RemoteAddr with the real
// TCP remote address. The golang.org/x/net/websocket package returns the
// WebSocket Origin header from RemoteAddr() on server-side connections,
// which is not the actual client IP. This wrapper fixes that by using the
// underlying HTTP request's RemoteAddr.
type wsRealAddrConn struct {
	net.Conn
	realAddr net.Addr
}

func (c *wsRealAddrConn) RemoteAddr() net.Addr {
	return c.realAddr
}

// httpRemoteAddr wraps the HTTP request's RemoteAddr string as a net.Addr.
type httpRemoteAddr string

func (a httpRemoteAddr) Network() string { return "tcp" }
func (a httpRemoteAddr) String() string  { return string(a) }

type WebsocketListener struct {
	ln       net.Listener
	acceptCh chan net.Conn

	server *http.Server
}

// NewWebsocketListener to handle websocket connections
// ln: tcp listener for websocket connections
func NewWebsocketListener(ln net.Listener) (wl *WebsocketListener) {
	wl = &WebsocketListener{
		ln:       ln,
		acceptCh: make(chan net.Conn),
	}

	muxer := http.NewServeMux()
	muxer.Handle(FrpWebsocketPath, websocket.Handler(func(c *websocket.Conn) {
		// The tunnel payload is a raw byte stream (yamux), not UTF-8 text.
		// Send it as binary frames; otherwise RFC 6455-compliant intermediaries
		// (e.g. API gateways/reverse proxies) UTF-8-validate the default text
		// frames and close the connection on invalid bytes.
		c.PayloadType = websocket.BinaryFrame

		// websocket.Conn.RemoteAddr() returns the WebSocket Origin header
		// (e.g. "http://host:port") on server-side connections, not the real
		// TCP client address. Use the HTTP request's RemoteAddr instead.
		var conn net.Conn = c
		if req := c.Request(); req != nil && req.RemoteAddr != "" {
			conn = &wsRealAddrConn{Conn: c, realAddr: httpRemoteAddr(req.RemoteAddr)}
		}

		notifyCh := make(chan struct{})
		wrapped := WrapCloseNotifyConn(conn, func(_ error) {
			close(notifyCh)
		})
		wl.acceptCh <- wrapped
		<-notifyCh
	}))

	wl.server = &http.Server{
		Addr:              ln.Addr().String(),
		Handler:           muxer,
		ReadHeaderTimeout: 60 * time.Second,
	}

	go func() {
		_ = wl.server.Serve(ln)
	}()
	return
}

func (p *WebsocketListener) Accept() (net.Conn, error) {
	c, ok := <-p.acceptCh
	if !ok {
		return nil, ErrWebsocketListenerClosed
	}
	return c, nil
}

func (p *WebsocketListener) Close() error {
	return p.server.Close()
}

func (p *WebsocketListener) Addr() net.Addr {
	return p.ln.Addr()
}
