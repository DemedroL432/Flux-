package multiserver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"pulseshield/app/config"
	"pulseshield/app/models"
)

const CacheTTL = 24 * time.Hour
const cacheFileName = "servers_cache.json"

type Cache struct {
	SubscriptionURL string               `json:"subscriptionUrl"`
	FetchedAt       time.Time            `json:"fetchedAt"`
	Servers         []models.VlessServer `json:"servers"`
}

func CachePath() (string, error) {
	dataDir, err := config.EnsureDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, cacheFileName), nil
}

func LoadCache(subscriptionURL string) (*Cache, bool, error) {
	path, err := CachePath()
	if err != nil {
		return nil, false, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var cache Cache
	if err := json.Unmarshal(raw, &cache); err != nil {
		return nil, false, err
	}
	if cache.SubscriptionURL != subscriptionURL {
		return &cache, false, nil
	}
	if time.Since(cache.FetchedAt) > CacheTTL {
		return &cache, false, nil
	}
	return &cache, true, nil
}

func SaveCache(subscriptionURL string, servers []models.VlessServer) error {
	path, err := CachePath()
	if err != nil {
		return err
	}
	cache := Cache{
		SubscriptionURL: subscriptionURL,
		FetchedAt:       time.Now(),
		Servers:         servers,
	}
	raw, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
