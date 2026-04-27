package correlation

import (
	"fmt"
	"strings"
)

// GenerateSummary creates a human-readable synthesis of cluster intelligence
func GenerateSummary(c Cluster) IdentitySummary {
	var platforms []string
	platformMap := make(map[string]bool)
	for _, m := range c.Members {
		if !platformMap[m.Platform] {
			platformMap[m.Platform] = true
			platforms = append(platforms, m.Platform)
		}
	}

	startYear := 0
	endYear := 0
	if len(c.Timeline.Events) > 0 {
		startYear = c.Timeline.Events[0].Timestamp.Year()
		endYear = c.Timeline.Events[len(c.Timeline.Events)-1].Timestamp.Year()
	}

	// 1. Infer Persona Type
	persona := "individual"
	if platformMap["GitHub"] {
		persona = "developer/security researcher"
	} else if platformMap["Twitter"] || platformMap["Reddit"] {
		persona = "social media user"
	}

	// 2. Behavioral Patterns
	pattern := "regular"
	if c.BehaviorProfile.ActivityPattern == "night_active" {
		pattern = "late-night"
	}

	// 3. Geographic Context
	geoStr := "global"
	if c.BehaviorProfile.Geo.ProbableRegion != "Global/Undetermined" {
		geoStr = c.BehaviorProfile.Geo.ProbableRegion
	}

	// 4. Construct Narrative Summary
	narrative := fmt.Sprintf("This identity appears to belong to a %s active across %s", persona, strings.Join(platforms, " and "))
	if startYear != 0 {
		narrative += fmt.Sprintf(", with consistent activity observed between %d–%d.", startYear, endYear)
	} else {
		narrative += "."
	}
	narrative += fmt.Sprintf(" Behavior suggests a %s activity pattern, likely centered in the %s region. Correlation is rated as %s.",
		pattern, geoStr, strings.ToLower(c.ConfidenceLevel))

	// 5. Key Points
	keyPoints := []string{}
	if len(platforms) > 1 {
		keyPoints = append(keyPoints, fmt.Sprintf("Multi-platform presence verified across %d sources", len(platforms)))
	}
	if c.BehaviorProfile.ConsistencyScore > 0.8 {
		keyPoints = append(keyPoints, "High behavioral consistency across mapped accounts")
	}
	for _, reason := range c.Reasons {
		keyPoints = append(keyPoints, "Signal: "+reason)
	}

	// 6. Risk Flags
	riskFlags := []string{}
	if c.Confidence < 0.75 {
		riskFlags = append(riskFlags, "Low-to-medium confidence linkage")
	}
	if len(c.BehaviorProfile.Anomalies) > 0 {
		riskFlags = append(riskFlags, "Behavioral anomalies detected in activity history")
	}
	if c.UncertaintyFlag {
		riskFlags = append(riskFlags, "UNCERTAINTY: Weak signal convergence")
	}
	if c.BehaviorProfile.AnomalyScore > 0.5 { // Example threshold for high anomaly risk
		riskFlags = append(riskFlags, fmt.Sprintf("HIGH ANOMALY SCORE (%.2f)", c.BehaviorProfile.AnomalyScore))
	}

	evidence := c.Reasons

	return IdentitySummary{
		Summary:      narrative,
		KeyPoints:    keyPoints,
		RiskFlags:    riskFlags,
		Confidence:   c.ConfidenceLevel,
		Evidence:     evidence,
		Score:        c.Confidence,
		AnomalyScore: c.BehaviorProfile.AnomalyScore, // Assign the new field
	}
}
