package external

import (
	"fmt"
	"nocturne/scanner/internal/models"
	"time"
)

// ExternalPlugin is a stub for an API-based data source
type ExternalPlugin struct{}

// NewPlugin creates a new instance of the external plugin
func NewPlugin() *ExternalPlugin {
	return &ExternalPlugin{}
}

// Name returns the identifier for this plugin
func (p *ExternalPlugin) Name() string {
	return "external_api"
}

// Run simulates an API-based search
func (p *ExternalPlugin) Run(input string) ([]models.Result, error) {
	// Simulate network latency
	time.Sleep(1 * time.Second)

	// Return a stub result
	return []models.Result{
		{
			Platform:   "ExternalDB",
			URL:        fmt.Sprintf("https://api.external.com/v1/search?q=%s", input),
			Exists:     true,
			Confidence: 0.9,
			Source:     p.Name(),
		},
	}, nil
}
