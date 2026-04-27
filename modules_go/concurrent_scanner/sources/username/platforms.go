package username

// DetectionRule defines how to identify if a user exists or not on a platform
type DetectionRule struct {
	Type          string // "status_code" or "body_contains"
	Value         string // The status code (as string) or the string to look for in the body
	ExpectExists  bool   // Whether finding this rule means the user exists or doesn't
}

// Platform represents a social media or web platform to scan
type Platform struct {
	Name           string
	URLFormat      string // e.g., "https://github.com/%s"
	DetectionRules []DetectionRule
}

// GetDefaultPlatforms returns a list of pre-configured platforms
func GetDefaultPlatforms() []Platform {
	return []Platform{
		{
			Name:      "GitHub",
			URLFormat: "https://github.com/%s",
			DetectionRules: []DetectionRule{
				{Type: "status_code", Value: "200", ExpectExists: true},
				{Type: "status_code", Value: "404", ExpectExists: false},
			},
		},
		{
			Name:      "Twitter",
			URLFormat: "https://twitter.com/%s",
			DetectionRules: []DetectionRule{
				{Type: "status_code", Value: "200", ExpectExists: true},
				{Type: "status_code", Value: "404", ExpectExists: false},
				{Type: "body_contains", Value: "This account doesn’t exist", ExpectExists: false},
			},
		},
		{
			Name:      "Reddit",
			URLFormat: "https://www.reddit.com/user/%s",
			DetectionRules: []DetectionRule{
				{Type: "status_code", Value: "200", ExpectExists: true},
				{Type: "status_code", Value: "404", ExpectExists: false},
				{Type: "body_contains", Value: "Sorry, nobody on Reddit goes by that name", ExpectExists: false},
			},
		},
		{
			Name:      "Instagram",
			URLFormat: "https://www.instagram.com/%s/",
			DetectionRules: []DetectionRule{
				{Type: "status_code", Value: "200", ExpectExists: true},
				{Type: "status_code", Value: "404", ExpectExists: false},
			},
		},
	}
}
