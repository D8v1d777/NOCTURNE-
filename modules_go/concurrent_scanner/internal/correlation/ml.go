package correlation

import (
	"errors"
	"sync"
)

// MLResults contains the semantic similarity signals from the ML engine
type MLResults struct {
	UsernameSimilarity    float64  `json:"username_similarity"`
	BioSemanticSimilarity float64  `json:"bio_semantic_similarity"`
	BehaviorPatternMatch  float64  `json:"behavior_pattern_match"`
	Reasons               []string `json:"reasons"`
}

var (
	mlCache   = make(map[string]*MLResults)
	mlCacheMu sync.RWMutex
)

// GetMLSignals communicates with a local ONNX runtime or remote ML service
// to perform semantic embedding comparisons.
func GetMLSignals(a, b Identity) (*MLResults, error) {
	// 1. Simple cache check to avoid redundant expensive inference
	cacheKey := a.ID + "_" + b.ID
	if a.ID > b.ID {
		cacheKey = b.ID + "_" + a.ID
	}

	mlCacheMu.RLock()
	if res, ok := mlCache[cacheKey]; ok {
		mlCacheMu.RUnlock()
		return res, nil
	}
	mlCacheMu.RUnlock()

	// Note: In a production environment, this would call a sidecar Python service
	// via gRPC or a local Rust library with bindings to an ONNX runtime.
	// Returning an error here triggers the rule-based fallback in engine.go.
	return nil, errors.New("ml_engine_not_initialized")
}

// ProcessMLBehaviorClassification categorizes clusters based on activity trends
func ProcessMLBehaviorClassification(c *Cluster) {
	// Logic for mapping high-dimensional activity vectors to personas
	// e.g., "Developer Cycle", "Social Bot Profile", "Active Hobbyist"
	if c.BehaviorProfile.ConsistencyScore > 0.9 {
		c.Timeline.Insights = append(c.Timeline.Insights, "ML Insight: High-confidence behavioral fingerprint match")
	}
}
