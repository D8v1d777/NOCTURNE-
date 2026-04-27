package username

import (
	"nocturne/scanner/internal/models"
)

// UsernamePlugin wraps the Scanner to implement the models.Source interface
type UsernamePlugin struct {
	scanner *Scanner
}

// NewPlugin creates a new instance of the username plugin
func NewPlugin() *UsernamePlugin {
	return &UsernamePlugin{
		scanner: NewScanner(),
	}
}

// Name returns the identifier for this plugin
func (p *UsernamePlugin) Name() string {
	return "username_scanner"
}

// Run executes the username scan
func (p *UsernamePlugin) Run(input string) ([]models.Result, error) {
	// The input is expected to be the username
	results := p.scanner.ScanUsername(input)
	return results, nil
}
