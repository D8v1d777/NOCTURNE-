package correlation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// RustSimilarityResult matches the JSON output from the Rust binary
type RustSimilarityResult struct {
	IDA     string   `json:"id_a"`
	IDB     string   `json:"id_b"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
}

// CallRustEngine offloads identity comparison to the Rust binary
func CallRustEngine(identities []Identity) ([]RustSimilarityResult, error) {
	// 1. Prepare input
	input, err := json.Marshal(identities)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal identities: %w", err)
	}

	// 2. Set up execution context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 3. Execute Rust binary
	// Note: Path is hardcoded for demo. In production, this would be configurable.
	cmd := exec.CommandContext(ctx, "./similarity_engine.exe") // .exe for windows
	cmd.Stdin = bytes.NewReader(input)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("rust engine error: %w (stderr: %s)", err, stderr.String())
	}

	// 4. Parse output
	var results []RustSimilarityResult
	err = json.Unmarshal(stdout.Bytes(), &results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rust output: %w", err)
	}

	return results, nil
}
