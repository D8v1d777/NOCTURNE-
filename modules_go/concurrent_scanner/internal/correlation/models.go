package correlation

import "time"

// Identity represents a raw OSINT profile found on a platform
type Identity struct {
	ID          string            `json:"id"`
	Platform    string            `json:"platform"`
	Username    string            `json:"username"`
	DisplayName string            `json:"display_name"`
	Bio         string            `json:"bio"`
	AvatarURL   string            `json:"avatar_url"`
	Links       []string          `json:"links"`
	Metadata    map[string]string `json:"metadata"`

	// Normalized fields for comparison (internal use)
	NormalizedUsername string              `json:"-"`
	BioTokens          map[string]struct{} `json:"-"`
	AvatarHash         string              `json:"-"`
	LinkDomains        map[string]struct{} `json:"-"`
}

type IdentitySummary struct {
	Summary      string   `json:"summary"`
	KeyPoints    []string `json:"key_points"`
	RiskFlags    []string `json:"risk_flags"`
	Confidence   string   `json:"confidence"`
	Evidence     []string `json:"evidence"`
	Score        float64  `json:"score"`
	AnomalyScore float64  `json:"anomaly_score"`
}

// Cluster represents a group of identities that are likely the same person
type Cluster struct {
	ID                string          `json:"id"`
	Members           []Identity      `json:"members"`
	Confidence        float64         `json:"confidence"`
	ConfidenceLevel   string          `json:"confidence_level"`
	ConfidenceExplain string          `json:"confidence_explanation"`
	Reasons           []string        `json:"reasons"`
	Timeline          Timeline        `json:"timeline"`
	BehaviorProfile   BehaviorProfile `json:"behavior_profile"`
	UncertaintyFlag   bool            `json:"uncertainty_flag"`
	Summary           IdentitySummary `json:"summary"`
}

// IdentityChange captures specific deltas for the Diff Viewer/Timeline UI
type IdentityChange struct {
	Type      string    `json:"type"` // e.g., "bio_update", "link_added"
	Field     string    `json:"field"`
	Old       string    `json:"old"`
	New       string    `json:"new"`
	Timestamp time.Time `json:"timestamp"`
	Platform  string    `json:"platform"`
	DiffHTML  string    `json:"diff_html,omitempty"` // For highlighting changes in UI
}
