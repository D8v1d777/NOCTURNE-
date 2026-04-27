package correlation

import (
	"strings"
)

// Stopwords list for lightweight NLP preprocessing
var stopwords = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "and": {}, "or": {}, "but": {}, "if": {}, "then": {},
	"else": {}, "when": {}, "where": {}, "why": {}, "how": {}, "is": {}, "are": {},
	"was": {}, "were": {}, "be": {}, "been": {}, "being": {}, "in": {}, "on": {},
	"at": {}, "to": {}, "from": {}, "by": {}, "for": {}, "with": {}, "about": {},
	"into": {}, "through": {}, "during": {}, "before": {}, "after": {}, "above": {},
	"below": {}, "of": {}, "up": {}, "down": {}, "out": {}, "off": {}, "over": {},
	"under": {}, "again": {}, "further": {}, "once": {}, "here": {}, "there": {},
	"all": {}, "any": {}, "both": {}, "each": {}, "few": {}, "more": {}, "most": {},
	"other": {}, "some": {}, "such": {}, "no": {}, "nor": {}, "not": {}, "only": {},
	"own": {}, "same": {}, "so": {}, "than": {}, "too": {}, "very": {}, "s": {},
	"t": {}, "can": {}, "will": {}, "just": {}, "don": {}, "should": {}, "now": {},
	"i": {}, "me": {}, "my": {}, "myself": {}, "we": {}, "our": {}, "ours": {},
	"ourselves": {}, "you": {}, "your": {}, "yours": {}, "yourself": {},
	"yourselves": {}, "he": {}, "him": {}, "his": {}, "himself": {}, "she": {},
	"her": {}, "hers": {}, "herself": {}, "it": {}, "its": {}, "itself": {},
	"they": {}, "them": {}, "their": {}, "theirs": {}, "themselves": {},
}

// TokenizeBio processes a bio string into meaningful tokens
func TokenizeBio(bio string) map[string]struct{} {
	tokens := make(map[string]struct{})
	
	// Lowercase and split into fields
	words := strings.Fields(strings.ToLower(bio))
	
	for _, word := range words {
		// Clean the word (keep @ handles and # hashtags as they are strong signals)
		clean := cleanWord(word)
		
		// Skip empty, short words, or stopwords
		if len(clean) <= 2 {
			continue
		}
		if _, isStopword := stopwords[clean]; isStopword {
			continue
		}
		
		tokens[clean] = struct{}{}
	}
	
	return tokens
}

// cleanWord removes common punctuation but preserves @ and #
func cleanWord(w string) string {
	return strings.TrimFunc(w, func(r rune) bool {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '@' || r == '#' {
			return false
		}
		return true
	})
}

// CalculateJaccard calculates the Jaccard similarity between two token sets
func CalculateJaccard(set1, set2 map[string]struct{}) float64 {
	if len(set1) == 0 || len(set2) == 0 {
		return 0.0
	}
	
	intersection := 0
	for k := range set1 {
		if _, exists := set2[k]; exists {
			intersection++
		}
	}
	
	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(union)
}

// GetSharedKeywords returns the overlapping tokens between two sets
func GetSharedKeywords(set1, set2 map[string]struct{}) []string {
	var shared []string
	for k := range set1 {
		if _, exists := set2[k]; exists {
			shared = append(shared, k)
		}
	}
	return shared
}
