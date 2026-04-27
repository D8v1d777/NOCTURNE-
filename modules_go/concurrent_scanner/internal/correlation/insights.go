package correlation

import (
	"fmt"
)

// GenerateStrategicInsights synthesizes high-level findings from all modules
func GenerateStrategicInsights(c Cluster) []StrategicInsight {
	var insights []StrategicInsight

	// 1. Stability Analysis
	stabilityMsg := "Identity shows high behavioral consistency and predictable activity cycles."
	stabilityConf := c.BehaviorProfile.ConsistencyScore
	if len(c.Timeline.History) > 3 {
		stabilityMsg = "Identity exhibits frequent rebranding or profile shifts, reducing long-term stability."
		stabilityConf *= 0.7
	}
	insights = append(insights, StrategicInsight{
		Category:   "Stability",
		Message:    stabilityMsg,
		Confidence: stabilityConf,
	})

	// 2. Risk Analysis
	if c.BehaviorProfile.AnomalyScore > 0.4 || c.UncertaintyFlag {
		riskLevel := "Medium"
		if c.BehaviorProfile.AnomalyScore > 0.7 {
			riskLevel = "High"
		}
		insights = append(insights, StrategicInsight{
			Category:   "Risk",
			Message:    fmt.Sprintf("%s risk detected: Recent anomalies indicate potential account expansion or behavioral drift.", riskLevel),
			Confidence: c.BehaviorProfile.AnomalyScore,
		})
	} else {
		insights = append(insights, StrategicInsight{
			Category:   "Risk",
			Message:    "Low risk profile: No significant behavioral deviations observed in current period.",
			Confidence: 0.9,
		})
	}

	// 3. Growth & Expansion
	platformCount := len(c.BehaviorProfile.PlatformUsage)
	if platformCount > 3 {
		insights = append(insights, StrategicInsight{
			Category:   "Growth",
			Message:    fmt.Sprintf("Broad digital footprint: Identity established across %d distinct platforms.", platformCount),
			Confidence: 1.0,
		})
	}

	for _, pred := range c.Timeline.Predictions {
		if pred.Type == "platform_expansion" && pred.Probability > 0.6 {
			insights = append(insights, StrategicInsight{
				Category:   "Growth",
				Message:    "High probability of expansion into new social or development platforms based on historical trends.",
				Confidence: pred.Probability,
			})
		}
	}

	// 4. Activity Pattern & Geo Alignment
	patternMsg := "Activity aligns with standard daytime cycles for inferred region."
	if c.BehaviorProfile.ActivityPattern == "night_active" {
		patternMsg = "Consistent 'Night-Active' pattern detected; typical of security research or hobbyist developers."
	}

	geoMsg := ""
	if c.BehaviorProfile.Geo.Confidence > 0.6 {
		geoMsg = fmt.Sprintf(" Strong alignment with %s timezone.", c.BehaviorProfile.Geo.ProbableRegion)
	}

	insights = append(insights, StrategicInsight{
		Category:   "Activity",
		Message:    patternMsg + geoMsg,
		Confidence: 0.85,
	})

	// 5. Correlation Integrity
	if c.Confidence > 0.85 {
		insights = append(insights, StrategicInsight{
			Category:   "Stability",
			Message:    "Cross-platform presence verified with high-confidence cryptographic and behavioral linkage.",
			Confidence: c.Confidence,
		})
	}

	return insights
}
