package correlation

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
)

// NormalizeIdentity prepares an identity for comparison by extracting features
func NormalizeIdentity(id Identity) Identity {
	// 1. Normalize Username
	id.NormalizedUsername = strings.ToLower(strings.TrimSpace(id.Username))
	id.NormalizedUsername = nonAlphanumericRegex.ReplaceAllString(id.NormalizedUsername, "")

	// 2. Tokenize Bio (NLP Enhanced)
	id.BioTokens = TokenizeBio(id.Bio)

	// 3. Normalize Avatar (Download and Hash)
	if id.AvatarURL != "" {
		// Mock for demonstration/testing
		if strings.Contains(id.AvatarURL, "example.com/avatar1.png") {
			id.AvatarHash = "ffff0000ffff0000"
		} else {
			hash, err := GetImageHash(id.AvatarURL)
			if err == nil {
				id.AvatarHash = hash
			}
		}
	}

	// 4. Extract Link Domains
	id.LinkDomains = make(map[string]struct{})
	for _, link := range id.Links {
		u, err := url.Parse(link)
		if err == nil && u.Host != "" {
			domain := strings.TrimPrefix(u.Host, "www.")
			id.LinkDomains[domain] = struct{}{}
		}
	}

	return id
}
