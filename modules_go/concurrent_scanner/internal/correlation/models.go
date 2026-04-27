package correlation

// Identity represents a raw OSINT profile found on a platform
type Identity struct {
	ID          string   `json:"id"`
	Platform    string   `json:"platform"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name"`
	Bio         string   `json:"bio"`
	AvatarURL   string   `json:"avatar_url"`
	Links       []string `json:"links"`

	// Normalized fields for comparison (internal use)
	NormalizedUsername string              `json:"-"`
	BioTokens          map[string]struct{} `json:"-"`
	AvatarHash         string              `json:"-"`
	LinkDomains        map[string]struct{} `json:"-"`
}

// Cluster represents a group of identities that are likely the same person
type Cluster struct {
	Members           []Identity `json:"members"`
	Confidence        float64    `json:"confidence"`
	ConfidenceLevel   string     `json:"confidence_level"`
	ConfidenceExplain string     `json:"confidence_explanation"`
	Reasons           []string   `json:"reasons"` // Why these identities were linked
}
