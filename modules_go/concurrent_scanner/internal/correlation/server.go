package api

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"nocturne/scanner/internal/correlation"
	"strconv"
	"strings"
	"sync"
)

// AnalysisCache stores results to prevent recomputing expensive correlation logic
type AnalysisCache struct {
	mu    sync.RWMutex
	store map[string]*CachedResult
}

type CachedResult struct {
	Clusters         []correlation.Cluster
	GraphEdges       []correlation.Edge // Flat list of all edges for graph visualization
	PrimaryClusterID string             // ID of the cluster relevant to the target query
	Expiry           int64
}

var Cache = &AnalysisCache{store: make(map[string]*CachedResult)} // Exported for scheduler access

// RegisterHandlers sets up the API routes for the frontend bridge
func RegisterHandlers() {
	// Wrap handlers with Gzip compression for performance
	http.HandleFunc("/api/graph", gzipWrapper(getGraphHandler))
	http.HandleFunc("/api/timeline", gzipWrapper(getTimelineHandler))
	http.HandleFunc("/api/behavior", gzipWrapper(getBehaviorHandler))
}

func gzipWrapper(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		fn(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func getGraphHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	target := query.Get("target")
	if target == "" {
		http.Error(w, "missing target parameter", http.StatusBadRequest)
		return
	}

	Cache.mu.RLock()
	res, exists := Cache.store[target]
	Cache.mu.RUnlock()

	// Ensure cache hit and that data hasn't expired
	if !exists || (res.Expiry > 0 && res.Expiry < strings.Index(target, "now")) { // simplified expiry check
		http.Error(w, "no analysis found for target", http.StatusNotFound)
		return
	}

	// Transform internal cluster structures into the format expected by Cytoscape.js
	type Node struct {
		ID         string  `json:"id"`
		Label      string  `json:"label"`
		Platform   string  `json:"platform"`
		Confidence float64 `json:"confidence"`
		Bio        string  `json:"bio"`
	}
	type Edge struct {
		Source string  `json:"source"` // Node ID
		Target string  `json:"target"` // Node ID
		Weight float64 `json:"weight"`
		Reason string  `json:"reason"`
	}

	graph := struct {
		Nodes []Node `json:"nodes"`
		Edges []Edge `json:"edges"`
	}{Nodes: []Node{}, Edges: []Edge{}}

	// Implement Pagination for massive clusters
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	if limit == 0 {
		limit = 1000 // Default safety cap
	}

	allIdentities := correlation.FlattenIdentities(res.Clusters)
	endNode := offset + limit
	if endNode > len(allIdentities) {
		endNode = len(allIdentities)
	}

	for i := offset; i < endNode; i++ {
		member := allIdentities[i]
		// Determine which cluster this member belongs to for confidence
		var conf float64
		for _, c := range res.Clusters {
			for _, m := range c.Members {
				if m.ID == member.ID {
					conf = c.Confidence
					break
				}
			}
		}
		graph.Nodes = append(graph.Nodes, Node{
			ID:         member.ID,
			Label:      member.Username,
			Platform:   member.Platform,
			Confidence: conf,
			Bio:        member.Bio,
		})
	}

	// Populate graph edges from cached GraphEdges
	idMap := make(map[int]string)
	for i, identity := range allIdentities {
		idMap[i] = identity.ID
	}

	// Edge pagination/filtering
	for _, edge := range res.GraphEdges {
		if edge.Weight < 0.6 { // Collapse low-confidence edges by default on backend
			continue
		}
		graph.Edges = append(graph.Edges, Edge{
			Source: idMap[edge.From],
			Target: idMap[edge.To],
			Weight: edge.Weight,
			Reason: strings.Join(edge.Reasons, ", "),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(graph)
}

func getTimelineHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	Cache.mu.RLock()
	res, exists := Cache.store[target]
	Cache.mu.RUnlock()

	if !exists {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res.Timeline)
}

func getBehaviorHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	Cache.mu.RLock()
	res, _ := Cache.store[target]
	Cache.mu.RUnlock()

	json.NewEncoder(w).Encode(res.Timeline.Profile)
}
