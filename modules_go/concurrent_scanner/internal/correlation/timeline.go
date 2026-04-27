package correlation

import (
	"fmt"
	"sort"
	"time"
)

// EventType represents the category of a temporal signal
type EventType string

const (
	AccountCreated EventType = "account_created"
	PostActivity   EventType = "post_activity"
	ProfileUpdate  EventType = "profile_update"
	LinkFound      EventType = "link_found"
)

// TimelineEvent represents a single point in time for an identity
type TimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        EventType `json:"type"`
	Description string    `json:"description"`
	Source      string    `json:"source"` // e.g., "Twitter", "GitHub"
	IsAnomaly   bool      `json:"is_anomaly,omitempty"`
}

type StrategicInsight struct {
	Category   string  `json:"category"` // Stability, Risk, Growth, Activity
	Message    string  `json:"message"`
	Confidence float64 `json:"confidence"`
}

// Timeline represents the full chronological history of a cluster
type Timeline struct {
	Events            []TimelineEvent    `json:"events"`
	Insights          []string           `json:"insights"`
	Profile           BehaviorProfile    `json:"profile"`
	Summary           IdentitySummary    `json:"summary"`
	Predictions       []Prediction       `json:"predictions"`
	History           []IdentityChange   `json:"history"`
	StrategicInsights []StrategicInsight `json:"strategic_insights"`
}

type IdentityChange struct {
	Type      string    `json:"type"`
	Old       string    `json:"old"`
	New       string    `json:"new"`
	Timestamp time.Time `json:"timestamp"`
	Platform  string    `json:"platform"`
}

type Prediction struct {
	Type          string    `json:"type"`
	Probability   float64   `json:"probability"`
	EstimatedTime time.Time `json:"estimated_time,omitempty"`
	Reason        string    `json:"reason"`
}

type BehaviorProfile struct {
	ActivityPattern  string         `json:"activity_pattern"`
	ConsistencyScore float64        `json:"consistency_score"`
	PlatformUsage    map[string]int `json:"platform_usage"`
	Anomalies        []string       `json:"anomalies"`
	Geo              GeoInference   `json:"geo"`
	AnomalyScore     float64        `json:"anomaly_score"` // New field
}

type GeoInference struct {
	ProbableTimezone string   `json:"probable_timezone"`
	ProbableRegion   string   `json:"probable_region"`
	Confidence       float64  `json:"confidence"`
	Reasoning        []string `json:"reasoning"`
}

// GenerateTimeline aggregates events from all identities in a cluster
func GenerateTimeline(identities []Identity) Timeline {
	var events []TimelineEvent

	for _, id := range identities {
		// 1. Extract creation date (if available in metadata)
		if createdStr, ok := id.Metadata["created_at"]; ok {
			if t, err := time.Parse(time.RFC3339, createdStr); err == nil {
				events = append(events, TimelineEvent{
					Timestamp:   t,
					Type:        AccountCreated,
					Description: fmt.Sprintf("Account created on %s", id.Platform),
					Source:      id.Platform,
				})
			}
		}

		// 2. Extract last activity
		if lastSeenStr, ok := id.Metadata["last_seen"]; ok {
			if t, err := time.Parse(time.RFC3339, lastSeenStr); err == nil {
				events = append(events, TimelineEvent{
					Timestamp:   t,
					Type:        PostActivity,
					Description: fmt.Sprintf("Last observed activity on %s", id.Platform),
					Source:      id.Platform,
				})
			}
		}

		// 3. Extract profile links as events
		for _, link := range id.Links {
			events = append(events, TimelineEvent{
				Timestamp:   time.Now(), // Default to current if unknown, or extract from page metadata
				Type:        LinkFound,
				Description: fmt.Sprintf("Cross-platform link discovered: %s", link),
				Source:      id.Platform,
			})
		}
	}

	// Sort events chronologically
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	insights, profile := analyzeTimeline(events, identities)

	return Timeline{
		Events:      events,
		Insights:    insights,
		Profile:     profile,
		Predictions: generatePredictions(events, profile),
	}
}

func analyzeTimeline(events []TimelineEvent, identities []Identity) ([]string, BehaviorProfile) {
	var insights []string
	profile := BehaviorProfile{
		PlatformUsage: make(map[string]int),
		Anomalies:     []string{},
	}

	if len(events) < 2 {
		return insights, profile
	}

	anomalyScore := 0.0
	// anomalyCount := 0 // Not strictly needed if we use weighted sum

	first := events[0]
	last := events[len(events)-1]
	hourlyDist := make(map[int]int)

	// 1. First Appearance Insight
	insights = append(insights, fmt.Sprintf("Identity first appeared in %d on %s", first.Timestamp.Year(), first.Source))

	// 2. Platform Usage & Hourly Distribution
	for i := range events {
		profile.PlatformUsage[events[i].Source]++
		hourlyDist[events[i].Timestamp.Hour()]++
	}

	// 3. Activity Pattern (Night vs Day)
	nightEvents := 0
	for h, count := range hourlyDist {
		if h >= 22 || h <= 4 {
			nightEvents += count
		}
	}
	if float64(nightEvents)/float64(len(events)) > 0.6 {
		profile.ActivityPattern = "night_active"
		insights = append(insights, "User is active mostly between 10PM–4AM (UTC)")
	} else {
		profile.ActivityPattern = "standard_day"
	}

	// 4. Anomaly Detection: Bursts & Spikes
	// Contribution for a burst
	burstThreshold := 3 // 3 events in 30 days
	for i := 0; i < len(events)-burstThreshold+1; i++ {
		windowStart := events[i].Timestamp
		windowEnd := events[i+burstThreshold-1].Timestamp
		if windowEnd.Sub(windowStart) < (30 * 24 * time.Hour) {
			msg := fmt.Sprintf("High activity burst detected around %s %d", windowStart.Month(), windowStart.Year())
			insights = append(insights, msg)
			profile.Anomalies = append(profile.Anomalies, msg)
			for j := i; j < i+burstThreshold; j++ {
				events[j].IsAnomaly = true
			}
			anomalyScore += 0.3
		}
	}

	// 5. Dormancy & Behavior Shifts (Contribution for dormancy)
	for i := 0; i < len(events)-1; i++ {
		gap := events[i+1].Timestamp.Sub(events[i].Timestamp)
		if gap > (365 * 24 * time.Hour) {
			msg := fmt.Sprintf("Behavior shift: Identity was dormant for %.1f years between %d and %d",
				gap.Hours()/24/365, events[i].Timestamp.Year(), events[i+1].Timestamp.Year())
			insights = append(insights, msg)
			profile.Anomalies = append(profile.Anomalies, msg)
			events[i+1].IsAnomaly = true
			anomalyScore += 0.2
		}

		// Platform switching coordination
		if events[i].Source != events[i+1].Source && gap < (1*time.Hour) {
			insights = append(insights, fmt.Sprintf("Coordination: Rapid platform switch from %s to %s", events[i].Source, events[i+1].Source))
		}
	}

	// 6. Sudden Expansion Detection (Contribution for new platform after long time)
	firstPlatformTime := make(map[string]time.Time)
	for _, e := range events {
		if _, ok := firstPlatformTime[e.Source]; !ok {
			firstPlatformTime[e.Source] = e.Timestamp
			if e.Timestamp.Sub(first.Timestamp) > (2 * 365 * 24 * time.Hour) {
				msg := fmt.Sprintf("Expansion: New platform %s joined after long established presence", e.Source)
				insights = append(insights, msg)
				profile.Anomalies = append(profile.Anomalies, msg)
				anomalyScore += 0.25
			}
		}
	}

	// 7. Consistency Score
	profile.ConsistencyScore = 1.0 - (float64(len(profile.Anomalies)) * 0.1)
	if len(profile.PlatformUsage) > 1 {
		profile.ConsistencyScore += 0.1 // Coordination bonus
	}
	if profile.ConsistencyScore > 1.0 {
		profile.ConsistencyScore = 1.0
	} else if profile.ConsistencyScore < 0.2 {
		profile.ConsistencyScore = 0.2
	}

	// 8. Current Status
	if time.Since(last.Timestamp) > (180 * 24 * time.Hour) {
		insights = append(insights, "Identity appears inactive (no activity in > 6 months)")
	}

	// Run Geo Inference
	profile.Geo = InferGeo(events, identities, hourlyDist)

	// Normalize Anomaly Score (cap at 1.0)
	if anomalyScore > 1.0 {
		anomalyScore = 1.0
	}
	profile.AnomalyScore = anomalyScore

	return insights, profile
}

func generatePredictions(events []TimelineEvent, profile BehaviorProfile) []Prediction {
	var predictions []Prediction
	if len(events) < 3 {
		return predictions
	}

	last := events[len(events)-1]

	// 1. Next Activity Prediction
	// Analyze the last 5 events to determine frequency
	windowSize := 5
	if len(events) < windowSize {
		windowSize = len(events)
	}
	recentEvents := events[len(events)-windowSize:]

	var totalGap time.Duration
	var gaps []time.Duration
	for i := 1; i < len(recentEvents); i++ {
		gap := recentEvents[i].Timestamp.Sub(recentEvents[i-1].Timestamp)
		gaps = append(gaps, gap)
		totalGap += gap
	}

	if len(gaps) > 0 {
		avgGap := totalGap / time.Duration(len(gaps))

		// Heuristic: If gaps are somewhat consistent, predict next activity
		isConsistent := true
		for _, g := range gaps {
			if float64(g) > float64(avgGap)*1.8 || float64(g) < float64(avgGap)*0.4 {
				isConsistent = false
				break
			}
		}

		if isConsistent && avgGap < 45*24*time.Hour {
			prob := 0.55
			if profile.ConsistencyScore > 0.8 {
				prob = 0.82
			}
			predictions = append(predictions, Prediction{
				Type:          "next_activity",
				Probability:   prob,
				EstimatedTime: last.Timestamp.Add(avgGap),
				Reason:        fmt.Sprintf("Consistent activity detected every %v", avgGap.Truncate(time.Hour)),
			})
		}
	}

	// 2. Platform Expansion Prediction
	if len(profile.PlatformUsage) > 1 {
		prob := 0.35 + (profile.ConsistencyScore * 0.2)
		predictions = append(predictions, Prediction{
			Type:        "platform_expansion",
			Probability: prob,
			Reason:      "Historical data shows multi-platform expansion patterns",
		})
	}

	// 3. Reactivation Probability (for inactive targets)
	if time.Since(last.Timestamp) > 180*24*time.Hour {
		prob := 0.15
		if profile.ConsistencyScore > 0.7 {
			prob = 0.4
		}
		predictions = append(predictions, Prediction{
			Type:        "reactivation",
			Probability: prob,
			Reason:      "Analysis of historical dormancy and return cycles",
		})
	}

	return predictions
}
