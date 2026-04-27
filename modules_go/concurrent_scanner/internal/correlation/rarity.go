package correlation

import (
	"strings"
	"unicode"
)

var commonUsernames = map[string]struct{}{
	"admin": {}, "test": {}, "user": {}, "guest": {}, "root": {},
	"john": {}, "smith": {}, "doe": {}, "anon": {}, "system": {},
	"support": {}, "info": {}, "mail": {}, "webmaster": {},
	"official": {}, "personal": {}, "dev": {}, "staff": {}, "bot": {},
}

// GetUsernameRarity returns a factor between 0.5 (common) and 1.2 (rare)
func GetUsernameRarity(username string) float64 {
	username = strings.ToLower(username)

	// 1. Check if it's a known common name
	if _, common := commonUsernames[username]; common {
		return 0.3 // Significant reduction for generic accounts
	}

	// 2. Short usernames are less likely to be unique
	if len(username) <= 4 {
		return 0.4
	}

	// 3. Check for "entropy" (digits, special chars, length)
	rarity := 0.8

	// Length bonus
	if len(username) > 10 {
		rarity += 0.2
	}

	// Complexity bonus
	var hasDigits, hasSpecial bool
	for _, r := range username {
		if unicode.IsDigit(r) {
			hasDigits = true
		} else if !unicode.IsLetter(r) {
			hasSpecial = true
		}
		if hasDigits && hasSpecial {
			break
		}
	}

	if hasDigits && hasSpecial {
		rarity += 0.2
	} else if hasDigits || hasSpecial {
		rarity += 0.1
	}

	return rarity
}
