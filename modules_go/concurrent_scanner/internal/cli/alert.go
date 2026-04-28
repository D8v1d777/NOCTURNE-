package cli

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"nocturne/scanner/internal/correlation"
)

// AlertType defines the category of an alert
type AlertType string

const (
	NewAccountDetected    AlertType = "new_account_detected"
	ProfileChanged        AlertType = "profile_changed"
	NewLinkFound          AlertType = "new_link_found"
	NewPlatformAppearance AlertType = "new_platform_appearance"
	ActivitySpike         AlertType = "activity_spike" // Derived from BehaviorProfile anomalies
	BehaviorAnomaly       AlertType = "behavior_anomaly"
)

// AlertSeverity defines the urgency/importance of an alert
type AlertSeverity string

const (
	SeverityLow    AlertSeverity = "low"
	SeverityMedium AlertSeverity = "medium"
	SeverityHigh   AlertSeverity = "high"
)

// Alert represents a detected change or event
type Alert struct {
	Type      AlertType              `json:"type"`
	Severity  AlertSeverity          `json:"severity"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	TargetID  string                 `json:"target_id"`         // The ID of the correlated identity/cluster
	Details   map[string]interface{} `json:"details,omitempty"` // Specifics about the change
}

// AlertRule defines conditions for triggering an alert
type AlertRule struct {
	Type                AlertType
	MinConfidence       float64       // For cluster-level alerts, e.g., NewAccountDetected
	Platforms           []string      // Specific platforms to monitor, empty for all
	BehaviorAnomalyType string        // Specific anomaly from BehaviorProfile (e.g., "activity burst")
	Severity            AlertSeverity // Override default severity if rule matches
}

// ClusterSnapshot stores the full state of a correlated cluster for comparison
type ClusterSnapshot struct {
	Cluster  correlation.Cluster
	Timeline correlation.Timeline
	History  []correlation.IdentityChange
}

// AlertManager handles alert generation, deduplication, and output
type AlertManager struct {
	mu                sync.Mutex
	lastAlerts        map[string]time.Time // Key: alert hash/signature, Value: last triggered time
	cooldown          time.Duration
	logFile           *os.File
	rules             []AlertRule
	previousSnapshots map[string]ClusterSnapshot // Stores the last known state of a cluster
}

// NewAlertManager creates a new AlertManager instance
func NewAlertManager(cooldown time.Duration, logFilePath string, rules []AlertRule) (*AlertManager, error) {
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open alert log file: %w", err)
	}

	return &AlertManager{
		lastAlerts:        make(map[string]time.Time),
		cooldown:          cooldown,
		logFile:           file,
		rules:             rules,
		previousSnapshots: make(map[string]ClusterSnapshot),
	}, nil
}

// StartStreamConsumer begins listening for correlation events to process alerts
func (am *AlertManager) StartStreamConsumer(bus *correlation.StreamBus) {
	go func() {
		eventChan := bus.Subscribe(correlation.TopicCorrelationEvent)
		log.Println("AlertManager started consuming correlation_events")

		for event := range eventChan {
			cluster, ok := event.Payload.(*correlation.Cluster)
			if !ok {
				log.Printf("AlertManager error: invalid payload type for target %s", event.TargetID)
				continue
			}

			am.ProcessClusterChange(event.TargetID, cluster)
		}
	}()
}

// Close closes the alert log file
func (am *AlertManager) Close() error {
	if am.logFile != nil {
		return am.logFile.Close()
	}
	return nil
}

// ProcessClusterChange compares a current cluster state with its previous snapshot
// and generates alerts based on defined rules.
func (am *AlertManager) ProcessClusterChange(targetID string, currentCluster *correlation.Cluster) {
	am.mu.Lock()
	defer am.mu.Unlock()

	previousSnapshot, exists := am.previousSnapshots[targetID]
	previousCluster := previousSnapshot.Cluster
	previousTimeline := previousSnapshot.Timeline

	history := previousSnapshot.History

	// If no previous snapshot, this is the first observation
	if !exists {
		am.previousSnapshots[targetID] = ClusterSnapshot{Cluster: *currentCluster, Timeline: currentCluster.Timeline}
		am.emitInitialAlerts(targetID, *currentCluster)
		return
	}

	// Compare current and previous clusters to detect changes
	history = am.detectIdentityChanges(targetID, *currentCluster, previousCluster, history)
	am.detectBehavioralChanges(targetID, currentCluster.Timeline, previousTimeline)

	// Inject evolution history and identify patterns
	currentCluster.Timeline.History = history
	currentCluster.Timeline.Insights = append(currentCluster.Timeline.Insights, am.detectEvolutionInsights(history)...)

	// Update snapshot
	am.previousSnapshots[targetID] = ClusterSnapshot{
		Cluster:  *currentCluster,
		Timeline: currentCluster.Timeline,
		History:  history,
	}
}

func (am *AlertManager) emitInitialAlerts(targetID string, cluster correlation.Cluster) {
	// Alert for new cluster/identity discovery
	alert := Alert{
		Type:      NewAccountDetected,
		Severity:  SeverityHigh,
		Message:   fmt.Sprintf("New correlated identity cluster discovered for target '%s'", targetID),
		Timestamp: time.Now().UTC(),
		TargetID:  targetID,
		Details:   map[string]interface{}{"members_count": len(cluster.Members), "confidence": cluster.Confidence},
	}
	am.triggerAlert(alert)

	// Also check for new platform appearances in the initial scan
	platforms := make(map[string]struct{})
	for _, member := range cluster.Members {
		if _, ok := platforms[member.Platform]; !ok {
			platforms[member.Platform] = struct{}{}
			alert := Alert{
				Type:      NewPlatformAppearance,
				Severity:  SeverityMedium,
				Message:   fmt.Sprintf("New platform '%s' observed for target '%s'", member.Platform, targetID),
				Timestamp: time.Now().UTC(),
				TargetID:  targetID,
				Details:   map[string]interface{}{"platform": member.Platform, "username": member.Username},
			}
			am.triggerAlert(alert)
		}
	}
}

func (am *AlertManager) detectIdentityChanges(targetID string, current, previous correlation.Cluster, history []correlation.IdentityChange) []correlation.IdentityChange {
	prevMembers := make(map[string]correlation.Identity) // Key: Identity.ID
	for _, member := range previous.Members {
		prevMembers[member.ID] = member
	}

	for _, currentMember := range current.Members {
		if prevMember, ok := prevMembers[currentMember.ID]; ok {
			// Existing member, check for changes
			if currentMember.Bio != prevMember.Bio {
				history = append(history, correlation.IdentityChange{
					Type:      "bio_update",
					Old:       prevMember.Bio,
					New:       currentMember.Bio,
					Timestamp: time.Now().UTC(),
					Platform:  currentMember.Platform,
				})
				alert := Alert{
					Type:      ProfileChanged,
					Severity:  SeverityLow,
					Message:   fmt.Sprintf("Bio changed for '%s' on %s", currentMember.Username, currentMember.Platform),
					Timestamp: time.Now().UTC(),
					TargetID:  targetID,
					Details:   map[string]interface{}{"identity_id": currentMember.ID, "platform": currentMember.Platform, "old_bio": prevMember.Bio, "new_bio": currentMember.Bio},
				}
				am.triggerAlert(alert)
			}
			if currentMember.AvatarHash != prevMember.AvatarHash {
				history = append(history, correlation.IdentityChange{
					Type:      "avatar_change",
					Old:       prevMember.AvatarHash,
					New:       currentMember.AvatarHash,
					Timestamp: time.Now().UTC(),
					Platform:  currentMember.Platform,
				})
				alert := Alert{
					Type:      ProfileChanged,
					Severity:  SeverityMedium,
					Message:   fmt.Sprintf("Avatar changed for '%s' on %s", currentMember.Username, currentMember.Platform),
					Timestamp: time.Now().UTC(),
					TargetID:  targetID,
					Details:   map[string]interface{}{"identity_id": currentMember.ID, "platform": currentMember.Platform, "old_avatar_hash": prevMember.AvatarHash, "new_avatar_hash": currentMember.AvatarHash},
				}
				am.triggerAlert(alert)
			}
			// Check for new links
			newLinks := getNewLinks(currentMember.Links, prevMember.Links)
			if len(newLinks) > 0 {
				history = append(history, correlation.IdentityChange{
					Type:      "new_links",
					Old:       strings.Join(prevMember.Links, ", "),
					New:       strings.Join(currentMember.Links, ", "),
					Timestamp: time.Now().UTC(),
					Platform:  currentMember.Platform,
				})
				alert := Alert{
					Type:      NewLinkFound,
					Severity:  SeverityMedium,
					Message:   fmt.Sprintf("New links found for '%s' on %s", currentMember.Username, currentMember.Platform),
					Timestamp: time.Now().UTC(),
					TargetID:  targetID,
					Details:   map[string]interface{}{"identity_id": currentMember.ID, "platform": currentMember.Platform, "new_links": newLinks},
				}
				am.triggerAlert(alert)
			}
		} else {
			// Detect rebranding (username change on established platform)
			isRebrand := false
			for _, pm := range previous.Members {
				if pm.Platform == currentMember.Platform {
					// If previous member is gone but same cluster is maintained via other signals
					foundInCurrent := false
					for _, cm := range current.Members {
						if cm.ID == pm.ID {
							foundInCurrent = true
							break
						}
					}
					if !foundInCurrent {
						history = append(history, correlation.IdentityChange{
							Type:      "username_change",
							Old:       pm.Username,
							New:       currentMember.Username,
							Timestamp: time.Now().UTC(),
							Platform:  currentMember.Platform,
						})
						isRebrand = true
						break
					}
				}
			}

			if !isRebrand {
				history = append(history, correlation.IdentityChange{
					Type:      "platform_addition",
					New:       currentMember.Platform,
					Timestamp: time.Now().UTC(),
					Platform:  currentMember.Platform,
				})
			}

			alert := Alert{
				Type:      NewAccountDetected,
				Severity:  SeverityMedium,
				Message:   fmt.Sprintf("New account '%s' (%s) correlated to target '%s'", currentMember.Username, currentMember.Platform, targetID),
				Timestamp: time.Now().UTC(),
				TargetID:  targetID,
				Details:   map[string]interface{}{"identity_id": currentMember.ID, "platform": currentMember.Platform, "username": currentMember.Username},
			}
			am.triggerAlert(alert)
		}
	}

	// Check for new platforms in the overall cluster
	currentPlatforms := make(map[string]struct{})
	for _, member := range current.Members {
		currentPlatforms[member.Platform] = struct{}{}
	}
	previousPlatforms := make(map[string]struct{})
	for _, member := range previous.Members {
		previousPlatforms[member.Platform] = struct{}{}
	}

	for p := range currentPlatforms {
		if _, ok := previousPlatforms[p]; !ok {
			alert := Alert{
				Type:      NewPlatformAppearance,
				Severity:  SeverityMedium,
				Message:   fmt.Sprintf("New platform '%s' appeared in cluster for target '%s'", p, targetID),
				Timestamp: time.Now().UTC(),
				TargetID:  targetID,
				Details:   map[string]interface{}{"platform": p},
			}
			am.triggerAlert(alert)
		}
	}
	return history
}

func (am *AlertManager) detectEvolutionInsights(history []correlation.IdentityChange) []string {
	var insights []string
	rebrands := 0
	for _, change := range history {
		if change.Type == "username_change" {
			rebrands++
			insights = append(insights, fmt.Sprintf("Identity Shift: Rebranded from '%s' to '%s' on %s", change.Old, change.New, change.Platform))
		}
	}

	if rebrands > 1 {
		insights = append(insights, "Insight: Identity shows a pattern of frequent alias shifting/rebranding.")
	}
	return insights
}
func getNewLinks(currentLinks, previousLinks []string) []string {
	prevLinkMap := make(map[string]struct{})
	for _, link := range previousLinks {
		prevLinkMap[link] = struct{}{}
	}
	var newLinks []string
	for _, link := range currentLinks {
		if _, ok := prevLinkMap[link]; !ok {
			newLinks = append(newLinks, link)
		}
	}
	return newLinks
}

func (am *AlertManager) detectBehavioralChanges(targetID string, currentTimeline, previousTimeline correlation.Timeline) {
	// Check for new behavior anomalies
	for _, currentAnomaly := range currentTimeline.Profile.Anomalies {
		isNew := true
		for _, prevAnomaly := range previousTimeline.Profile.Anomalies {
			if currentAnomaly == prevAnomaly {
				isNew = false
				break
			}
		}
		if isNew {
			alert := Alert{
				Type:      BehaviorAnomaly,
				Severity:  SeverityMedium, // Default severity, can be overridden by rules
				Message:   fmt.Sprintf("New behavioral anomaly detected for target '%s': %s (Anomaly Score: %.2f)", targetID, currentAnomaly, currentTimeline.Profile.AnomalyScore),
				Timestamp: time.Now().UTC(),
				TargetID:  targetID,
				Details:   map[string]interface{}{"anomaly": currentAnomaly},
			}
			am.triggerAlert(alert)
		}
	}

	// Check for activity pattern change (e.g., from day_active to night_active)
	if previousTimeline.Profile.ActivityPattern != currentTimeline.Profile.ActivityPattern &&
		currentTimeline.Profile.ActivityPattern != "" && previousTimeline.Profile.ActivityPattern != "" {
		alert := Alert{
			Type:      BehaviorAnomaly, // Could also be a specific "ActivityPatternShift" type
			Severity:  SeverityLow,
			Message:   fmt.Sprintf("Activity pattern shift for target '%s': from '%s' to '%s' (Anomaly Score: %.2f)", targetID, previousTimeline.Profile.ActivityPattern, currentTimeline.Profile.ActivityPattern, currentTimeline.Profile.AnomalyScore),
			Timestamp: time.Now().UTC(),
			TargetID:  targetID,
			Details:   map[string]interface{}{"old_pattern": previousTimeline.Profile.ActivityPattern, "new_pattern": currentTimeline.Profile.ActivityPattern},
		}
		am.triggerAlert(alert)
	}
}

func (am *AlertManager) triggerAlert(alert Alert) {
	// Generate a simple signature for deduplication.
	// For more robust deduplication, consider hashing the entire alert struct or specific fields.
	alertSignature := fmt.Sprintf("%s-%s-%s", alert.Type, alert.TargetID, alert.Message)

	if am.isDuplicate(alertSignature) {
		return // Alert is on cooldown
	}

	// Apply rules to potentially override severity or filter
	finalSeverity := alert.Severity // Start with default severity
	for _, rule := range am.rules {
		if rule.Type == alert.Type {
			// Check platform filter
			if len(rule.Platforms) > 0 {
				alertPlatform, ok := alert.Details["platform"].(string)
				if !ok || !contains(rule.Platforms, alertPlatform) {
					continue // Rule doesn't apply due to platform mismatch
				}
			}

			// Check behavior anomaly type filter
			if rule.BehaviorAnomalyType != "" {
				alertAnomaly, ok := alert.Details["anomaly"].(string)
				if !ok || !strings.Contains(alertAnomaly, rule.BehaviorAnomalyType) {
					continue // Rule doesn't apply due to anomaly type mismatch
				}
			}

			// Check min confidence for cluster-level alerts
			if rule.MinConfidence > 0 {
				if alert.Type == NewAccountDetected { // Only apply to NewAccountDetected for now
					if clusterConfidence, ok := alert.Details["confidence"].(float64); ok {
						if clusterConfidence < rule.MinConfidence {
							continue // Rule doesn't apply, confidence too low
						}
					}
				}
			}

			// If all conditions met, apply rule's severity
			finalSeverity = rule.Severity
			break // Apply the first matching rule and exit
		}
	}
	alert.Severity = finalSeverity

	// Output to console
	fmt.Printf("🚨 ALERT [%s] (%s): %s (Target: %s)\n", alert.Severity, alert.Type, alert.Message, alert.TargetID)

	// Output to log file
	logEntry := fmt.Sprintf("[%s] [%s] [%s] Target: %s - %s\n", alert.Timestamp.Format(time.RFC3339), alert.Severity, alert.Type, alert.TargetID, alert.Message)
	if _, err := am.logFile.WriteString(logEntry); err != nil {
		log.Printf("Error writing alert to log file: %v", err)
	}

	// Update last alert time for deduplication
	am.lastAlerts[alertSignature] = time.Now()
}

func (am *AlertManager) isDuplicate(alertSignature string) bool {
	if lastTime, ok := am.lastAlerts[alertSignature]; ok {
		if time.Since(lastTime) < am.cooldown {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
