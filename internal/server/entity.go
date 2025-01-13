package server

import (
	"net/http"
	"time"

	"github.com/docker/go-connections/nat"
	"go.etcd.io/bbolt"
)

type GameServer struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	ContainerID  string                 `json:"containerId"`
	State        int                    `json:"state"`
	Image        string                 `json:"image"`
	HostIP       string                 `json:"hostIp"`
	ExposedPorts map[string]nat.PortSet `json:"exposedPorts"`
	HealthCheck  string                 `json:"healthCheck"`
	RestartCount int                    `json:"restartCount"`
	StartTime    time.Time              `json:"startTime"`
	FinishTime   time.Time              `json:"finishTime,omitempty"`
}

// Template represents a game server configuration
type Template struct {
	Name         string            `json:"name"`
	Image        string            `json:"image"`
	DefaultPorts map[string]string `json:"defaultPorts"` // port -> default host port
	HealthCheck  string            `json:"healthCheck,omitempty"`
}

// Catalog manages our game server entities
type Catalog struct {
	db *bbolt.DB
}

// OrchestratorClient represents your existing API client
type OrchestratorClient struct {
	baseURL string
	client  *http.Client
}

// Task matches your existing orchestrator API response
type Task struct {
	ID           string                 `json:"ID"`
	ContainerID  string                 `json:"ContainerID"`
	Name         string                 `json:"Name"`
	State        int                    `json:"State"`
	Image        string                 `json:"Image"`
	ExposedPorts map[string]nat.PortSet `json:"ExposedPorts"`
	HostPorts    nat.PortMap            `json:"HostPorts"`
	PortBindings map[string]string      `json:"PortBindings"`
	HealthCheck  string                 `json:"HealthCheck"`
	RestartCount int                    `json:"RestartCount"`
	StartTime    time.Time              `json:"StartTime"`
	FinishTime   time.Time              `json:"FinishTime"`
}
