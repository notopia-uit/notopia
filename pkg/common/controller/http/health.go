package http

import (
	"context"
	"sync"
	"time"
)

type StartupStatus string

const (
	StartupStatusStarting StartupStatus = "starting"
	StartupStatusStarted  StartupStatus = "started"
	StartupStatusFailed   StartupStatus = "failed"
)

type StartupResponse struct {
	Status         StartupStatus  `json:"status"`
	Message        string         `json:"message,omitempty"`
	AdditionalInfo map[string]any `json:"additional_info,omitempty"`
}

type ReadinessStatus string

const (
	ReadinessStatusNotReady ReadinessStatus = "not_ready"
	ReadinessStatusReady    ReadinessStatus = "ready"
)

type ReadinessResponse struct {
	Status     ReadinessStatus            `json:"status"`
	Message    string                     `json:"message,omitempty"`
	Components map[string]HealthComponent `json:"components,omitempty"`
}

type HealthComponent struct {
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
}

type HealthToCheckFunc func(ctx context.Context) error

type HealthManager struct {
	toCheckReady map[string]HealthToCheckFunc
	isStarted    bool
	isReady      bool
	components   map[string]HealthComponent
	mu           sync.RWMutex
	checkTimeout time.Duration
}

func NewHealthManager(
	toCheckReady map[string]HealthToCheckFunc,
) *HealthManager {
	componentNames := make([]string, 0, len(toCheckReady))
	for name := range toCheckReady {
		componentNames = append(componentNames, name)
	}
	components := make(map[string]HealthComponent)
	for _, name := range componentNames {
		components[name] = HealthComponent{
			Healthy: false,
			Message: "not checked yet",
		}
	}
	return &HealthManager{
		toCheckReady: toCheckReady,
		checkTimeout: 5 * time.Second,
		components:   components,
	}
}

func (hm *HealthManager) SetStartedUp() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.isStarted = true
}

func (hm *HealthManager) StartupHTTPHandler(ctx context.Context) StartupResponse {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if !hm.isStarted {
		return StartupResponse{
			Status: StartupStatusStarting,
		}
	}

	return StartupResponse{
		Status: StartupStatusStarted,
	}
}

func (hm *HealthManager) ReadinessHTTPHandler(ctx context.Context) ReadinessResponse {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	status := ReadinessStatusReady
	if !hm.isStarted || !hm.isReady {
		status = ReadinessStatusNotReady
	}
	return ReadinessResponse{
		Status:     status,
		Components: hm.components,
	}
}

func (hm *HealthManager) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hm.checkReadiness(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (hm *HealthManager) checkReadiness(parentCtx context.Context) {
	ctx, cancel := context.WithTimeout(parentCtx, hm.checkTimeout)
	defer cancel()

	var wg sync.WaitGroup
	componentsMu := sync.Mutex{}
	components := make(map[string]HealthComponent)
	allHealthy := true

	for name, checkFunc := range hm.toCheckReady {
		wg.Add(1)
		go func(componentName string, check HealthToCheckFunc) {
			defer wg.Done()
			err := check(ctx)
			heathy := err == nil

			componentsMu.Lock()
			components[componentName] = HealthComponent{
				Healthy: heathy,
				Message: err.Error(),
			}
			if !heathy {
				allHealthy = false
			}
			componentsMu.Unlock()
		}(name, checkFunc)
	}
	wg.Wait()

	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.isReady = allHealthy
	hm.components = components
}
