package http

type StartupStatus string

const (
	StartupStatusUnknown  StartupStatus = "unknown"
	StartupStatusStarting StartupStatus = "starting"
	StartupStatusStarted  StartupStatus = "started"
)

type StartupResponse struct {
	Status         StartupStatus  `json:"status"`
	Message        string         `json:"message,omitempty"`
	AdditionalInfo map[string]any `json:"additional_info,omitempty"`
}

type ReadinessStatus string

const (
	ReadinessStatusUnknown  ReadinessStatus = "unknown"
	ReadinessStatusNotReady ReadinessStatus = "not_ready"
	ReadinessStatusReady    ReadinessStatus = "ready"
)

type ReadinessResponse struct {
	Status         ReadinessStatus `json:"status"`
	Message        string          `json:"message,omitempty"`
	AdditionalInfo map[string]any  `json:"additional_info,omitempty"`
}

type Health interface {
	StartupCheck() StartupResponse
	ReadinessCheck() ReadinessResponse
}
