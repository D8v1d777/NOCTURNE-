package external

import (
	"context"
	"nocturne/scanner/internal/engine"
	"nocturne/scanner/internal/models"
	"time"
)

// RustModulePlugin is a wrapper for a future Rust-based high-performance module
type RustModulePlugin struct {
	Config engine.ExternalModuleConfig
}

func NewRustModulePlugin() *RustModulePlugin {
	return &RustModulePlugin{
		Config: engine.ExternalModuleConfig{
			Name:    "rust_scanner",
			Path:    "./tools/rust_module_bin", // Path to the future Rust binary
			Timeout: 5 * time.Second,
		},
	}
}

func (p *RustModulePlugin) Name() string {
	return "rust_bridge"
}

func (p *RustModulePlugin) Run(input string) ([]models.Result, error) {
	// Prepare input struct for JSON
	moduleInput := struct {
		Username string `json:"username"`
	}{
		Username: input,
	}

	// Prepare output slice
	var moduleOutput []models.Result

	// Execute external module via the engine's runner
	// Note: In a real production environment, we would check if the binary exists first.
	// For this preparation phase, we are just implementing the flow.
	err := engine.RunExternalModule(context.Background(), p.Config, moduleInput, &moduleOutput)
	if err != nil {
		// If binary doesn't exist yet, we'll return a helpful error or empty results
		return nil, err
	}

	return moduleOutput, nil
}
