package username

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"nocturne/scanner/internal/models"
	"strings"
	"sync"
	"time"
)

// Scanner handles the concurrent scanning process
type Scanner struct {
	Workers    int
	RateLimit  time.Duration
	Timeout    time.Duration
	UserAgent  string
	Platforms  []Platform
}

// NewScanner creates a new scanner instance with default values
func NewScanner() *Scanner {
	return &Scanner{
		Workers:   10,
		RateLimit: 500 * time.Millisecond,
		Timeout:   10 * time.Second,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 NocturneScanner/1.0",
		Platforms: GetDefaultPlatforms(),
	}
}

// ScanUsername checks a username across all configured platforms
func (s *Scanner) ScanUsername(username string) []models.Result {
	results := make([]models.Result, 0, len(s.Platforms))
	resultChan := make(chan models.Result, len(s.Platforms))
	platformChan := make(chan Platform, len(s.Platforms))

	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < s.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for platform := range platformChan {
				resultChan <- s.checkPlatform(username, platform)
				if s.RateLimit > 0 {
					time.Sleep(s.RateLimit)
				}
			}
		}()
	}

	// Feed platforms
	go func() {
		for _, p := range s.Platforms {
			platformChan <- p
		}
		close(platformChan)
	}()

	// Close result channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		results = append(results, res)
	}

	return results
}

func (s *Scanner) checkPlatform(username string, p Platform) models.Result {
	url := fmt.Sprintf(p.URLFormat, username)
	result := models.Result{
		Platform:   p.Name,
		URL:        url,
		Confidence: 0.0,
	}

	client := &http.Client{
		Timeout: s.Timeout,
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	req.Header.Set("User-Agent", s.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("network error: %v", err)
		return result
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Sprintf("failed to read body: %v", err)
		return result
	}
	body := string(bodyBytes)

	exists := false
	confidence := 0.0
	matches := 0

	for _, rule := range p.DetectionRules {
		switch rule.Type {
		case "status_code":
			if fmt.Sprintf("%d", resp.StatusCode) == rule.Value {
				exists = rule.ExpectExists
				confidence = 1.0
				matches++
			}
		case "body_contains":
			if strings.Contains(body, rule.Value) {
				exists = rule.ExpectExists
				confidence = 0.8
				matches++
			}
		}
	}

	if matches == 0 {
		if resp.StatusCode == http.StatusOK {
			exists = true
			confidence = 0.5
		} else if resp.StatusCode == http.StatusNotFound {
			exists = false
			confidence = 1.0
		}
	}

	result.Exists = exists
	result.Confidence = confidence

	return result
}
