package models

// Result represents the structured output of any search source
type Result struct {
	Platform   string  `json:"platform"`
	URL        string  `json:"url"`
	Exists     bool    `json:"exists"`
	Confidence float64 `json:"confidence"`
	Error      string  `json:"error,omitempty"`
	Source     string  `json:"source"` // Which plugin provided this result
}

// Source is the interface that all NOCTURNE plugins must implement
type Source interface {
	Name() string
	Run(input string) ([]Result, error)
}
