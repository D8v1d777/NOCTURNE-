package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Identity struct {
	ID         string   `json:"id"`
	Username   string   `json:"username"`
	Bio        string   `json:"bio"`
	AvatarHash string   `json:"avatar_hash"`
	Links      []string `json:"links"`
}

type SimilarityResult struct {
	IDA     string   `json:"id_a"`
	IDB     string   `json:"id_b"`
	Score   float64  `json:"score"`
	Reasons []string `json:"reasons"`
}

func main() {
	// Read input
	input, _ := io.ReadAll(os.Stdin)
	var identities []Identity
	json.Unmarshal(input, &identities)

	var results []SimilarityResult

	// Simple mock logic: if usernames are identical, they match
	for i := 0; i < len(identities); i++ {
		for j := i + 1; j < len(identities); j++ {
			if identities[i].Username == identities[j].Username {
				results = append(results, SimilarityResult{
					IDA:     identities[i].ID,
					IDB:     identities[j].ID,
					Score:   1.0,
					Reasons: []string{"exact username match (rust-mock)"},
				})
			}
		}
	}

	// Output
	output, _ := json.Marshal(results)
	fmt.Println(string(output))
}
