package correlation

import (
	"fmt"
	"strings"
)

var ccTLDToRegion = map[string]string{
	"uk": "United Kingdom",
	"de": "Germany",
	"in": "India",
	"jp": "Japan",
	"fr": "France",
	"br": "Brazil",
	"ru": "Russia",
	"cn": "China",
	"au": "Australia",
	"ca": "Canada",
	"it": "Italy",
	"es": "Spain",
}

var langIndicators = map[string]string{
	"namaste": "India",
	"hindi":   "India",
	"bonjour": "France/Francophone",
	"hola":    "Spain/LATAM",
	"privet":  "Russia/Eastern Europe",
	"ciao":    "Italy",
	"servus":  "Germany/Austria",
}

// InferGeo uses behavioral and metadata signals to estimate location
func InferGeo(events []TimelineEvent, identities []Identity, hourlyDist map[int]int) GeoInference {
	geo := GeoInference{
		ProbableTimezone: "Unknown",
		ProbableRegion:   "Global/Undetermined",
		Confidence:       0.1,
		Reasoning:        []string{},
	}

	if len(events) == 0 {
		return geo
	}

	// 1. Timezone Inference via Peak Hour Analysis
	// Heuristic: Peak social/coding activity often occurs around 8PM (20:00) local time
	peakUTC := -1
	maxActivity := 0
	for h, count := range hourlyDist {
		if count > maxActivity {
			maxActivity = count
			peakUTC = h
		}
	}

	if peakUTC != -1 {
		// Offset = TargetLocalTime (20) - CurrentUTCTime
		offset := 20 - peakUTC
		if offset > 12 {
			offset -= 24
		} else if offset < -12 {
			offset += 24
		}

		geo.ProbableTimezone = fmt.Sprintf("UTC%+d", offset)
		geo.Reasoning = append(geo.Reasoning, fmt.Sprintf("Peak activity observed at %02d:00 UTC (Estimated offset: %+d)", peakUTC, offset))
		geo.Confidence += 0.2
	}

	// 2. Region Inference via ccTLD (Links)
	foundTLDs := make(map[string]int)
	for _, id := range identities {
		for domain := range id.LinkDomains {
			parts := strings.Split(domain, ".")
			if len(parts) > 1 {
				tld := parts[len(parts)-1]
				if region, ok := ccTLDToRegion[tld]; ok {
					foundTLDs[region]++
				}
			}
		}
	}

	// 3. Region Inference via Language (Bios)
	foundLangs := make(map[string]int)
	for _, id := range identities {
		bioLower := strings.ToLower(id.Bio)
		for keyword, region := range langIndicators {
			if strings.Contains(bioLower, keyword) {
				foundLangs[region]++
			}
		}
	}

	// Combine Region Results
	if len(foundTLDs) > 0 {
		// Simplified: pick the most frequent
		for region := range foundTLDs {
			geo.ProbableRegion = region
			geo.Reasoning = append(geo.Reasoning, fmt.Sprintf("Domain usage suggests connection to %s", region))
			geo.Confidence += 0.3
			break
		}
	}

	if len(foundLangs) > 0 {
		for region := range foundLangs {
			if geo.ProbableRegion == "Global/Undetermined" {
				geo.ProbableRegion = region
			}
			geo.Reasoning = append(geo.Reasoning, fmt.Sprintf("Linguistic markers suggest %s region", region))
			geo.Confidence += 0.2
			break
		}
	}

	if geo.Confidence > 0.85 {
		geo.Confidence = 0.85 // Avoid overconfidence on heuristics
	}

	return geo
}
