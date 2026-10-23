//go:build windows

package licensing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func DeviceHash() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.READ)
	if err != nil {
		return "", fmt.Errorf("реестр MachineGuid: %w", err)
	}
	defer k.Close()
	guid, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return "", fmt.Errorf("MachineGuid: %w", err)
	}
	guid = strings.TrimSpace(strings.ToLower(guid))
	if guid == "" {
		return "", fmt.Errorf("пустой MachineGuid")
	}
	sum := sha256.Sum256([]byte(guid + "|pulseshield-device-v1"))
	return hex.EncodeToString(sum[:]), nil
}
