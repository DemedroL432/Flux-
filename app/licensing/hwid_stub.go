//go:build !windows

package licensing

import "fmt"

func DeviceHash() (string, error) {
	return "", fmt.Errorf("идентификатор устройства доступен только в сборке Windows")
}
