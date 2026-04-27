package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// ExternalModuleConfig defines the execution parameters for an external module
type ExternalModuleConfig struct {
	Name    string
	Path    string
	Timeout time.Duration
}

// RunExternalModule executes an external binary, passes input via stdin (JSON),
// and captures output via stdout (JSON).
func RunExternalModule(ctx context.Context, config ExternalModuleConfig, input interface{}, output interface{}) error {
	// Setup context with timeout
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// Prepare the command
	cmd := exec.CommandContext(ctx, config.Path)

	// Prepare stdin
	inputData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}
	cmd.Stdin = bytes.NewReader(inputData)

	// Prepare stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	err = cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("module %s timed out after %v", config.Name, config.Timeout)
		}
		return fmt.Errorf("module %s failed: %w (stderr: %s)", config.Name, err, stderr.String())
	}

	// Parse output
	if err := json.Unmarshal(stdout.Bytes(), output); err != nil {
		return fmt.Errorf("failed to unmarshal module %s output: %w (raw: %s)", config.Name, err, stdout.String())
	}

	return nil
}
