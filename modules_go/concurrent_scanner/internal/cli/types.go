package distribution

import (
	"nocturne/scanner/internal/models"
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusAssigned  TaskStatus = "assigned"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type ScanTask struct {
	ID       string     `json:"id"`
	Target   string     `json:"target"`
	Plugins  []string   `json:"plugins"`
	Status   TaskStatus `json:"status"`
	WorkerID string     `json:"worker_id"`
	Retries  int        `json:"retries"`
}

type WorkerInfo struct {
	ID       string    `json:"id"`
	Addr     string    `json:"addr"`
	LastSeen time.Time `json:"last_seen"`
	Status   string    `json:"status"` // "active", "dead"
}

type ScanResponse struct {
	TaskID  string          `json:"task_id"`
	Results []models.Result `json:"results"`
}
