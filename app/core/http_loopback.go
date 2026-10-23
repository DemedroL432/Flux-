package core

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func hostLooksLoopback(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// HTTPTransportNoProxyOnLoopback copies DefaultTransport and skips HTTP proxies for loopback targets.
// Windows may route 127.0.0.1 through the system proxy when TUN/system proxy is enabled.
func HTTPTransportNoProxyOnLoopback() http.RoundTripper {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.Proxy = func(req *http.Request) (*url.URL, error) {
		if req.URL != nil && hostLooksLoopback(req.URL.Hostname()) {
			return nil, nil
		}
		return http.ProxyFromEnvironment(req)
	}
	return t
}

// NewLoopbackAwareHTTPClient is used for local APIs (Clash on 127.0.0.1, etc.).
func NewLoopbackAwareHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: HTTPTransportNoProxyOnLoopback(),
	}
}
