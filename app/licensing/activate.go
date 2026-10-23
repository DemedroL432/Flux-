package licensing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ActivateRequest struct {
	Code string `json:"code"`
	HWID string `json:"hwid_hash"`
	App  string `json:"app_version"`
}

// ActivateResponse — ответ сервера активации (см. docs/LICENSING.md).
type ActivateResponse struct {
	Tier          string `json:"tier"`
	ValidUntilUnix int64 `json:"valid_until_unix"`
	Token        string `json:"token,omitempty"`
	Error        string `json:"error"`
}

func ActivateCode(baseURL, code, appVersion string) (*ActivateResponse, error) {
	if ep := strings.TrimSpace(os.Getenv("PULSESHIELD_LICENSE_API")); ep != "" {
		baseURL = ep
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	code = strings.TrimSpace(code)
	if baseURL == "" {
		return nil, fmt.Errorf(`укажите базовый URL в настройках (licenseApiBase) или переменную окружения PULSESHIELD_LICENSE_API`)
	}
	if code == "" {
		return nil, fmt.Errorf("введите код активации")
	}
	hwid, err := DeviceHash()
	if err != nil {
		return nil, fmt.Errorf("устройство: %w", err)
	}

	url := baseURL + "/v1/activate"
	body, err := json.Marshal(ActivateRequest{Code: code, HWID: hwid, App: appVersion})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 28 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var out ActivateResponse
	_ = json.Unmarshal(raw, &out)

	if resp.StatusCode != http.StatusOK {
		if strings.TrimSpace(out.Error) != "" {
			return nil, fmt.Errorf("%s", out.Error)
		}
		return nil, fmt.Errorf("сервер: %s — %s", resp.Status, string(bytes.TrimSpace(raw)))
	}
	if strings.TrimSpace(out.Error) != "" {
		return nil, fmt.Errorf("%s", out.Error)
	}
	return &out, nil
}
