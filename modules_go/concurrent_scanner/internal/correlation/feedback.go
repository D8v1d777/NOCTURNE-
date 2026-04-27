package correlation

import (
	"encoding/json"
	"os"
	"sync"
)

const feedbackFile = "correlation_feedback.json"

type FeedbackStore struct {
	SignalReliability map[string]float64 `json:"signal_reliability"` // Signal -> Adjustment Factor (0.5 to 1.5)
	CorrectClusters   int                `json:"correct_clusters"`
	IncorrectClusters int                `json:"incorrect_clusters"`
	mu                sync.RWMutex
}

var (
	globalFeedback *FeedbackStore
	once           sync.Once
)

func GetFeedbackStore() *FeedbackStore {
	once.Do(func() {
		globalFeedback = &FeedbackStore{
			SignalReliability: make(map[string]float64),
		}
		globalFeedback.load()
	})
	return globalFeedback
}

func (s *FeedbackStore) load() {
	data, err := os.ReadFile(feedbackFile)
	if err == nil {
		json.Unmarshal(data, s)
	}
	
	// Initialize defaults if empty
	s.mu.Lock()
	defer s.mu.Unlock()
	defaultSignals := []string{
		"exact username match", 
		"fuzzy username match", 
		"exact avatar hash match", 
		"perceptual avatar match",
		"exact shared link",
		"shared link domain",
		"smart bio match",
	}
	for _, sig := range defaultSignals {
		if _, exists := s.SignalReliability[sig]; !exists {
			s.SignalReliability[sig] = 1.0
		}
	}
}

func (s *FeedbackStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(feedbackFile, data, 0644)
}

// GetAdjustment returns the reliability factor for a signal (defaults to 1.0)
func (s *FeedbackStore) GetAdjustment(signalPrefix string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Find the best matching signal prefix
	for sig, val := range s.SignalReliability {
		if len(signalPrefix) >= len(sig) && signalPrefix[:len(sig)] == sig {
			return val
		}
	}
	return 1.0
}

// ProcessFeedback adjusts weights based on user input
func (s *FeedbackStore) ProcessFeedback(signals []string, isCorrect bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	adjustment := 0.05
	if !isCorrect {
		adjustment = -0.1
		s.IncorrectClusters++
	} else {
		s.CorrectClusters++
	}

	for _, signal := range signals {
		for sig := range s.SignalReliability {
			// If the signal starts with one of our tracked categories
			if len(signal) >= len(sig) && signal[:len(sig)] == sig {
				newVal := s.SignalReliability[sig] + adjustment
				// Clamp between 0.5 and 1.5
				if newVal < 0.5 { newVal = 0.5 }
				if newVal > 1.5 { newVal = 1.5 }
				s.SignalReliability[sig] = newVal
			}
		}
	}
}
