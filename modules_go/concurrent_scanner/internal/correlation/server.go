package correlation

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

// AnalysisCache stores results to prevent recomputing expensive correlation logic
type AnalysisCache struct {
	mu    sync.RWMutex
	store map[string]*CachedResult
}

type CachedResult struct {
	Clusters         []Cluster
	GraphEdges       []Edge // Flat list of all edges for graph visualization
	Timeline         Timeline
	PrimaryClusterID string // ID of the cluster relevant to the target query
	Expiry           int64
}

// Cache is the global singleton for storing and retrieving analysis results.
var Cache = AnalysisCache{
	store: make(map[string]*CachedResult),
}

// IngestRequest matches the Python scheduler's output format
type IngestRequest struct {
	TargetID string   `json:"target_id"`
	Identity Identity `json:"identity"`
}

func StartServer() {
	// Register Graph API
	http.HandleFunc("/api/graph", handleGraph)

	// Server config
	port := getEnv("PORT", "8080")

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      nil, // uses http.DefaultServeMux
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🕯️ NOCTURNE API running on http://localhost:%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}

func handleGraph(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		target := r.URL.Query().Get("target")
		Cache.mu.RLock()
		res, exists := Cache.store[target]
		Cache.mu.RUnlock()

		if !exists {
			http.Error(w, "Target not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(res)

	case http.MethodPost:
		var req IngestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 1. Retrieve current state or create new
		Cache.mu.Lock()
		current, exists := Cache.store[req.TargetID]

		var identities []Identity
		if exists {
			identities = FlattenIdentities(current.Clusters)
		}

		// 2. Add the new identity
		identities = append(identities, req.Identity)

		// 3. Re-run correlation to update the Graph and Clusters
		clusters, edges := RunCorrelation(identities)

		// 4. Update the Cache
		newResult := &CachedResult{
			Clusters:   clusters,
			GraphEdges: edges,
			Expiry:     time.Now().Add(24 * time.Hour).Unix(),
		}

		// Identify the primary cluster (highest confidence)
		if len(clusters) > 0 {
			best := clusters[0]
			for _, c := range clusters {
				if c.Confidence > best.Confidence {
					best = c
				}
			}
			newResult.PrimaryClusterID = best.ID
			newResult.Timeline = best.Timeline
		}

		Cache.store[req.TargetID] = newResult
		Cache.mu.Unlock()

		log.Printf("Updated Graph for target [%s]: %d identities, %d clusters", req.TargetID, len(identities), len(clusters))
		w.WriteHeader(http.StatusAccepted)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// =======================
// UTIL
// =======================

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
