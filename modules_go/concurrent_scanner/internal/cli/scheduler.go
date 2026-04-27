package monitoring

import (
	"context"
	"log"
	"nocturne/scanner/internal/api"
	"nocturne/scanner/internal/engine"
	"nocturne/scanner/internal/events"
	"sync"
	"time"
)

// MonitoringScheduler orchestrates periodic scans and feeds results to the AlertManager
type MonitoringScheduler struct {
	alertManager  *AlertManager
	scanner       *engine.Manager    // The main scanning engine
	apiCache      *api.AnalysisCache // The cache to store and retrieve previous states
	bus           *events.StreamBus
	interval      time.Duration
	maxInterval   time.Duration // For adaptive scheduling
	targets       []string      // List of targets (e.g., usernames) to monitor
	targetBackoff map[string]time.Duration
	stopChan      chan struct{}
	wg            sync.WaitGroup
	workerPool    chan struct{} // Semaphore for concurrency control
}

// NewMonitoringScheduler creates a new scheduler
func NewMonitoringScheduler(am *AlertManager, scanner *engine.Manager, apiCache *api.AnalysisCache, bus *events.StreamBus, interval time.Duration, targets []string) *MonitoringScheduler {
	return &MonitoringScheduler{
		alertManager:  am,
		scanner:       scanner,
		apiCache:      apiCache,
		bus:           bus,
		interval:      interval,
		maxInterval:   interval * 10,
		targets:       targets,
		targetBackoff: make(map[string]time.Duration),
		stopChan:      make(chan struct{}),
		workerPool:    make(chan struct{}, 10), // Limit to 10 concurrent scans
	}
}

// Start begins the periodic monitoring process
func (ms *MonitoringScheduler) Start(ctx context.Context) {
	log.Printf("Monitoring scheduler started for targets: %v, interval: %s", ms.targets, ms.interval)
	ticker := time.NewTicker(ms.interval)
	defer ticker.Stop()

	ms.wg.Add(1)
	go func() {
		defer ms.wg.Done()
		for {
			select {
			case <-ticker.C:
				ms.runMonitoringCycle()
			case <-ms.stopChan:
				log.Println("Monitoring scheduler stopped.")
				return
			case <-ctx.Done():
				log.Println("Monitoring scheduler received context cancellation.")
				return
			}
		}
	}()
}

// Stop halts the monitoring process
func (ms *MonitoringScheduler) Stop() {
	close(ms.stopChan)
	ms.wg.Wait()
}

func (ms *MonitoringScheduler) runMonitoringCycle() {
	log.Println("Running monitoring cycle...")
	for _, target := range ms.targets {
		// Adaptive check: Skip if the backoff period hasn't elapsed
		if backoff, ok := ms.targetBackoff[target]; ok && backoff > 0 {
			// Logic to decrement or check elapsed time would go here
		}

		select {
		case ms.workerPool <- struct{}{}: // Acquire worker
			ms.wg.Add(1)
			go func(t string) {
				defer ms.wg.Done()
				defer func() { <-ms.workerPool }() // Release worker
				ms.monitorTarget(t)
			}(target)
		case <-ms.stopChan:
			return
		}
	}
	ms.wg.Wait()
}

func (ms *MonitoringScheduler) monitorTarget(target string) {
	log.Printf("Monitoring target: %s", target)

	// 1. Run initial scan
	scanResults := ms.scanner.RunPlugins(target, []string{"username_scanner", "external_api"})

	if len(scanResults) == 0 {
		return
	}

	// 2. Produce to Stream: Push raw data into the pipeline
	ms.bus.Publish(events.TopicRawData, target, scanResults)
	log.Printf("Published raw_data for %s to stream", target)
}
