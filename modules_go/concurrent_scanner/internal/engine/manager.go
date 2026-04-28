package engine

import (
	"log"
	"nocturne/scanner/internal/models"
	"sync"
)

// Manager handles the registration and execution of plugins
type Manager struct {
	plugins map[string]models.Source
	mu      sync.RWMutex
}

// NewManager creates a new plugin manager
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]models.Source),
	}
}

// Register adds a plugin to the manager
func (m *Manager) Register(p models.Source) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plugins[p.Name()] = p
}

// RunPlugins executes all enabled plugins concurrently
func (m *Manager) RunPlugins(input string, enabled []string) []models.Result {
	var wg sync.WaitGroup
	resultChan := make(chan []models.Result, len(enabled))

	enabledMap := make(map[string]bool)
	for _, name := range enabled {
		enabledMap[name] = true
	}

	m.mu.RLock()
	for name, plugin := range m.plugins {
		if !enabledMap[name] {
			continue
		}

		wg.Add(1)
		go func(p models.Source) {
			defer wg.Done()

			// Protect the main engine from plugin crashes
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Plugin %s panicked: %v", p.Name(), r)
				}
			}()

			results, err := p.Run(input)
			if err != nil {
				log.Printf("Plugin %s returned error: %v", p.Name(), err)
				// Even on error, we might return partial results or just the error result
				resultChan <- []models.Result{{
					Source: p.Name(),
					Error:  err.Error(),
				}}
				return
			}

			// Tag results with their source
			for i := range results {
				results[i].Source = p.Name()
			}
			resultChan <- results
		}(plugin)
	}
	m.mu.RUnlock()

	// Close channel when all plugins are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var finalResults []models.Result
	for res := range resultChan {
		finalResults = append(finalResults, res...)
	}

	return finalResults
}

// GetAvailablePlugins returns a list of registered plugin names
func (m *Manager) GetAvailablePlugins() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var names []string
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}
