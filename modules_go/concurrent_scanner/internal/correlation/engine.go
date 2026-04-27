package correlation

import (
	"fmt"
	"strings"
)

// Edge represents a similarity link between two identities
type Edge struct {
	From    int // Index of the source identity in the normalized slice
	To      int
	Weight  float64
	Reasons []string
}

// Compare calculates a similarity score between two identities
func Compare(a, b Identity) (float64, []string, []string) {
	var rejectionReasons []string

	// 1. Signal Gating (Early Exit)
	// Reject if:
	// - Username similarity < 0.5 AND
	// - No shared links AND
	// - No avatar match
	uSim := LevenshteinDistance(a.NormalizedUsername, b.NormalizedUsername)

	hasSharedLink := false
	for la := range a.Links {
		if _, exists := b.Links[la]; exists {
			hasSharedLink = true
			break
		}
	}
	if !hasSharedLink {
		for la := range a.Links {
			da := extractDomain(la)
			if da == "" {
				continue
			}
			for lb := range b.Links {
				db := extractDomain(lb)
				if da == db {
					hasSharedLink = true
					break
				}
			}
			if hasSharedLink {
				break
			}
		}
	}

	hasAvatarMatch := false
	if a.AvatarHash != "" && b.AvatarHash != "" {
		dist := HammingDistance(a.AvatarHash, b.AvatarHash)
		if dist <= 4 {
			hasAvatarMatch = true
		}
	}

	if uSim < 0.5 && !hasSharedLink && !hasAvatarMatch {
		return 0.0, nil, []string{"weak signals: rejected by gating logic"}
	}

	// 2. Strong Signal Requirement
	// Only allow scoring if at least ONE strong signal exists:
	// - Exact username match
	// - Avatar similarity
	// - Shared external link
	isExactUsername := (a.NormalizedUsername != "" && a.NormalizedUsername == b.NormalizedUsername)

	if !isExactUsername && !hasAvatarMatch && !hasSharedLink {
		return 0.0, nil, []string{"no strong signal (username/avatar/links)"}
	}

	score := 0.0
	var reasons []string
	hasStrongSignal := isExactUsername || hasAvatarMatch || hasSharedLink

	// 1. Username Scoring (Hybrid ML + Rule-based)
	rarity := GetUsernameRarity(a.Username)
	feedback := GetFeedbackStore()

	mlResults, mlErr := GetMLSignals(a, b)

	if isExactUsername {
		weight := 0.5 * rarity * feedback.GetAdjustment("exact username match")
		score += weight
		reasons = append(reasons, fmt.Sprintf("exact username match (rarity: %.1f)", rarity))
	} else {
		// Use ML for fuzzy username similarity if available
		uMatchScore := uSim
		if mlErr == nil && mlResults.UsernameSimilarity > uSim {
			uMatchScore = mlResults.UsernameSimilarity
		}

		if uMatchScore >= 0.85 {
			weight := 0.2 * rarity * feedback.GetAdjustment("fuzzy username match")
			score += weight
			if mlErr == nil && mlResults.UsernameSimilarity > uSim {
				reasons = append(reasons, fmt.Sprintf("ML-enhanced username match (sim: %.2f)", uMatchScore))
				reasons = append(reasons, mlResults.Reasons...)
			} else {
				reasons = append(reasons, fmt.Sprintf("fuzzy username match (sim: %.2f)", uSim))
			}
		} else {
			score -= 0.3
			rejectionReasons = append(rejectionReasons, "conflicting usernames")
		}
	}

	// 2. Avatar Similarity Scoring (Dynamic)
	if a.AvatarHash != "" && b.AvatarHash != "" {
		dist := HammingDistance(a.AvatarHash, b.AvatarHash)
		if dist == 0 {
			weight := 0.4 * feedback.GetAdjustment("exact avatar hash match")
			score += weight
			reasons = append(reasons, "exact avatar hash match")
		} else if dist <= 4 {
			weight := 0.25 * feedback.GetAdjustment("perceptual avatar match")
			score += weight
			reasons = append(reasons, "perceptual avatar match")
		}
	}

	// 3. Shared Links Scoring (Dynamic)
	if hasSharedLink {
		linkFound := false
		for la := range a.Links {
			if _, exists := b.Links[la]; exists {
				weight := 0.4 * feedback.GetAdjustment("exact shared link")
				score += weight
				reasons = append(reasons, "exact shared link: "+la)
				linkFound = true
				break
			}
		}
		if !linkFound {
			// Check domains since we know there's a domain match from gating
			for la := range a.Links {
				da := extractDomain(la)
				if da == "" {
					continue
				}
				for lb := range b.Links {
					db := extractDomain(lb)
					if da == db {
						weight := 0.15 * feedback.GetAdjustment("shared link domain")
						score += weight
						reasons = append(reasons, "shared link domain: "+da)
						linkFound = true
						break
					}
				}
				if linkFound {
					break
				}
			}
		}
	}

	// 4. Bio Similarity Scoring (Semantic ML Fallback)
	bioMatched := false
	if mlErr == nil && mlResults.BioSemanticSimilarity > 0.65 {
		bioWeight := 0.20 * mlResults.BioSemanticSimilarity * feedback.GetAdjustment("smart bio match")
		score += bioWeight
		reasons = append(reasons, fmt.Sprintf("ML semantic bio match (sim: %.2f)", mlResults.BioSemanticSimilarity))
		bioMatched = true
	} else {
		// Fallback to Jaccard
		jaccard := CalculateJaccard(a.BioTokens, b.BioTokens)
		if len(a.BioTokens) >= 3 && len(b.BioTokens) >= 3 && jaccard > 0.4 {
			bioWeight := 0.15
			if len(a.BioTokens) > 10 && len(b.BioTokens) > 10 {
				bioWeight = 0.25
			}
			bioWeight *= feedback.GetAdjustment("smart bio match")
			score += bioWeight
			shared := GetSharedKeywords(a.BioTokens, b.BioTokens)
			reasons = append(reasons, fmt.Sprintf("keyword bio match (weight: %.2f): %s", bioWeight, strings.Join(shared, ", ")))
			bioMatched = true
		}
	}

	if !bioMatched && len(a.BioTokens) > 5 && len(b.BioTokens) > 5 {
		// Strong penalty for different bios when username matches
		score -= 0.4
		rejectionReasons = append(rejectionReasons, "conflicting bio content")
	}

	// Final normalization: clamp between 0 and 1
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score, reasons, rejectionReasons
}

// RunCorrelation groups identities into clusters using graph-based connected components
func RunCorrelation(identities []Identity) ([]Cluster, []Edge) {
	if len(identities) == 0 {
		return nil, nil
	}

	// 1. Normalize all
	normalized := make([]Identity, len(identities))
	for i, id := range identities {
		normalized[i] = NormalizeIdentity(id)
	}

	// 2. Build Adjacency List (Try Rust first, fallback to Go)
	var allGraphEdges []Edge // To collect all unique edges for the graph visualization
	adj := make([][]Edge, len(normalized))

	rustResults, err := CallRustEngine(normalized)
	if err == nil && len(rustResults) > 0 {
		// Use high-performance Rust results
		idToIndex := make(map[string]int)
		for i, id := range normalized {
			idToIndex[id.ID] = i
		}

		for _, res := range rustResults {
			i := idToIndex[res.IDA]
			j := idToIndex[res.IDB]
			edge1 := Edge{From: i, To: j, Weight: res.Score, Reasons: res.Reasons}
			edge2 := Edge{From: j, To: i, Weight: res.Score, Reasons: res.Reasons}
			adj[i] = append(adj[i], edge1)
			adj[j] = append(adj[j], edge2)
			allGraphEdges = append(allGraphEdges, edge1)
		}
	} else {
		// Fallback to Go implementation
		// Use a higher threshold for conservative matching
		threshold := 0.70
		for i := 0; i < len(normalized); i++ {
			for j := i + 1; j < len(normalized); j++ {
				score, reasons, _ := Compare(normalized[i], normalized[j])
				if score >= threshold {
					edge1 := Edge{From: i, To: j, Weight: score, Reasons: reasons}
					edge2 := Edge{From: j, To: i, Weight: score, Reasons: reasons}
					adj[i] = append(adj[i], edge1)
					adj[j] = append(adj[j], edge2)
					allGraphEdges = append(allGraphEdges, edge1) // Store only one direction for graph
				}
			}
		}
	}

	// 3. Find Connected Components
	visited := make([]bool, len(normalized))
	var clusters []Cluster

	for i := 0; i < len(normalized); i++ {
		if !visited[i] {
			// New component found
			var componentIndices []int
			// var componentEdges []Edge // Edges within this component, if needed for cluster-specific graph
			stack := []int{i}
			visited[i] = true

			for len(stack) > 0 {
				curr := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				componentIndices = append(componentIndices, curr)

				for _, edge := range adj[curr] {
					if !visited[edge.To] {
						visited[edge.To] = true
						// componentEdges = append(componentEdges, edge)
						stack = append(stack, edge.To)
					}
				}
			}

			// 4. Create Cluster from component
			cluster := Cluster{
				ID:      fmt.Sprintf("cluster-%d", i), // Assign a unique ID to the cluster
				Members: make([]Identity, 0, len(componentIndices)),
			}

			totalWeight := 0.0
			// Note: edgeCount here counts each edge twice (once for From, once for To)
			// when iterating through adj[idx]. If we want unique edges, we need to adjust.
			// For confidence calculation, this is fine as it's about total connection strength.
			edgeCount := 0
			uniqueReasons := make(map[string]struct{})
			strongEdgeCount := 0

			for _, idx := range componentIndices {
				cluster.Members = append(cluster.Members, normalized[idx])
				for _, edge := range adj[idx] {
					totalWeight += edge.Weight
					edgeCount++
					if edge.Weight >= 0.8 {
						strongEdgeCount++
					}
					for _, r := range edge.Reasons {
						uniqueReasons[r] = struct{}{}
					}
				}
			}

			// Generate Timeline and BehaviorProfile for the cluster
			clusterTimeline := GenerateTimeline(cluster.Members)
			cluster.Timeline = clusterTimeline
			cluster.BehaviorProfile = clusterTimeline.Profile
			// 5. Confidence Calibration
			if len(componentIndices) > 1 && edgeCount > 0 {
				// Each undirected edge was counted twice
				avgSimilarity := totalWeight / float64(edgeCount)
				confidence := avgSimilarity

				// Penalty: Too many members with low average similarity
				if len(componentIndices) > 3 && avgSimilarity < 0.75 {
					confidence -= 0.1
				}

				// Penalty: Low strong edge density
				// edgeCount/2 is the actual number of unique edges
				strongEdgeDensity := float64(strongEdgeCount) / float64(edgeCount)
				if strongEdgeDensity < 0.3 {
					confidence -= 0.15
				}

				// Boost: High strong edge density
				if strongEdgeDensity > 0.7 {
					confidence += 0.1
				}

				// Boost: Consistent signals across members
				if len(uniqueReasons) >= 3 {
					confidence += 0.05
				}

				// Clamp and Level Assignment
				if confidence < 0 {
					confidence = 0
				}
				if confidence > 1 {
					confidence = 1
				}

				cluster.Confidence = confidence
				if confidence < 0.7 {
					cluster.UncertaintyFlag = true
				}
				switch {
				case confidence >= 0.9:
					cluster.ConfidenceLevel = "Very High"
				case confidence >= 0.75:
					cluster.ConfidenceLevel = "High"
				case confidence >= 0.6:
					cluster.ConfidenceLevel = "Medium"
				default:
					cluster.ConfidenceLevel = "Low (Rejected)"
				}

				// Build Explanation
				reasonsList := make([]string, 0, len(uniqueReasons))
				for r := range uniqueReasons {
					reasonsList = append(reasonsList, r)
				}
				cluster.Reasons = reasonsList
				cluster.ConfidenceExplain = fmt.Sprintf("matched %s with %.0f%% strong connections",
					strings.Join(reasonsList, " + "), strongEdgeDensity*100)

				// Generate strategic insights before summary
				cluster.Timeline.StrategicInsights = GenerateStrategicInsights(cluster)

				// Generate rule-based identity summary
				cluster.Summary = GenerateSummary(cluster)
				cluster.Timeline.Summary = cluster.Summary
			} else {
				cluster.Confidence = 1.0
				cluster.ConfidenceLevel = "N/A (Single Node)"
				cluster.ConfidenceExplain = "no comparison data available"
			}

			// Only add cluster if confidence is high enough
			if cluster.Confidence >= 0.6 || len(cluster.Members) == 1 {
				clusters = append(clusters, cluster)
			}
		}
	}

	return clusters, allGraphEdges
}

// Helper functions

func LevenshteinDistance(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	if s1 == "" || s2 == "" {
		return 0.0
	}

	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, min(d[i][j-1]+1, d[i-1][j-1]+cost))
		}
	}

	dist := d[len(s1)][len(s2)]
	maxLen := max(len(s1), len(s2))
	return 1.0 - float64(dist)/float64(maxLen)
}

func extractDomain(url string) string {
	if url == "" {
		return ""
	}
	// Simple domain extraction
	parts := strings.Split(url, "/")
	if len(parts) < 3 {
		return ""
	}
	domain := parts[2]
	// Remove port if present
	if strings.Contains(domain, ":") {
		domain = strings.Split(domain, ":")[0]
	}
	// Remove www.
	return strings.TrimPrefix(domain, "www.")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func countOverlap(set1, set2 map[string]struct{}) int {
	count := 0
	for key := range set1 {
		if _, exists := set2[key]; exists {
			count++
		}
	}
	return count
}

// FlattenIdentities returns a flat slice of all identities from a slice of clusters.
// This is useful for mapping internal edge indices to global Identity IDs.
func FlattenIdentities(clusters []Cluster) []Identity {
	var allIdentities []Identity
	for _, cluster := range clusters {
		allIdentities = append(allIdentities, cluster.Members...)
	}
	return allIdentities
}
