package state

import (
	"sync"

	"pulseshield/app/models"
)

type AppState struct {
	mu     sync.RWMutex
	config models.AppConfig
}

func New(initial models.AppConfig) *AppState {
	return &AppState{config: initial}
}

func (s *AppState) Snapshot() models.AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *AppState) Update(mutator func(cfg *models.AppConfig)) models.AppConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	mutator(&s.config)
	return s.config
}
