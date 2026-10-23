package winproxy

import (
	"strings"

	"golang.org/x/sys/windows/registry"
)

const internetSettingsKey = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

type Snapshot struct {
	ProxyEnable uint32
	ProxyServer string
	ProxyOverride string
}

func Read() (Snapshot, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.QUERY_VALUE)
	if err != nil {
		return Snapshot{}, err
	}
	defer k.Close()

	var snap Snapshot
	if v, _, err := k.GetIntegerValue("ProxyEnable"); err == nil {
		snap.ProxyEnable = uint32(v)
	}
	if v, _, err := k.GetStringValue("ProxyServer"); err == nil {
		snap.ProxyServer = v
	}
	if v, _, err := k.GetStringValue("ProxyOverride"); err == nil {
		snap.ProxyOverride = v
	}
	return snap, nil
}

func ApplyHTTPProxy(hostport, username, password string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	// Bare "host:port" sets a single HTTP proxy (via CONNECT) for ALL protocols.
	// We deliberately do NOT use the "http=…;https=…" form: the "https=" scheme
	// prefix makes some clients (Firefox/Chromium WinINET path) treat the proxy as
	// a *secure* (TLS) proxy and attempt a TLS handshake to 127.0.0.1:8080, which the
	// sing-box mixed inbound (plain HTTP CONNECT) rejects → "secure connection failed".
	// Auth is never embedded — the local inbound is unauthenticated (127.0.0.1 only).
	_ = username
	_ = password
	proxyServer := strings.TrimSpace(hostport)
	if err := k.SetDWordValue("ProxyEnable", 1); err != nil {
		return err
	}
	if err := k.SetStringValue("ProxyServer", proxyServer); err != nil {
		return err
	}
	_ = k.SetStringValue("ProxyOverride", "<local>")
	NotifyProxyChange()
	return nil
}

func Disable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.SetDWordValue("ProxyEnable", 0); err != nil {
		return err
	}
	NotifyProxyChange()
	return nil
}

