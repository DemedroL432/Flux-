package config

import (
	crand "crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"pulseshield/app/models"
)

// EnsureLocalInboundSecrets задаёт логин/пароль локального mixed inbound и Clash API при первом запуске.
func EnsureLocalInboundSecrets(cfg *models.AppConfig) bool {
	if cfg == nil {
		return false
	}
	changed := false
	if strings.TrimSpace(cfg.LocalInboundUser) == "" {
		cfg.LocalInboundUser = models.DefaultLocalInboundUser
		changed = true
	}
	if strings.TrimSpace(cfg.LocalInboundPassword) == "" {
		cfg.LocalInboundPassword = randomSecretHex(24)
		changed = true
	}
	return changed
}

func randomSecretHex(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := crand.Read(b); err != nil {
		// crypto/rand недоступен (крайне редко) — используем time-based псевдоэнтропию.
		// Результат не криптографически стойкий, но лучше предсказуемых нулей.
		t := time.Now().UnixNano()
		for i := range b {
			b[i] = byte(t>>uint(i%8)) ^ byte(i*137+42)
		}
	}
	return hex.EncodeToString(b)
}
