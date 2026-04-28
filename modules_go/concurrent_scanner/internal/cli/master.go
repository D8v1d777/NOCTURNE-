package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Master struct {
	mu            sync.RWMutex
	TaskQueue     []*ScanTask
	Workers       map[string]*WorkerInfo
	AssignedTasks map[string]*ScanTask // TaskID -> Task
}

func NewMaster() *Master {
	m := &Master{
		Workers:       make(map[string]*WorkerInfo),
		AssignedTasks: make(map[string]*ScanTask),
	}
	go m.startReaper()
	return m
}

// RegisterHandlers sets up Master API
func (m *Master) RegisterHandlers() {
	http.HandleFunc("/master/register", m.handleRegister)
	http.HandleFunc("/master/tasks/poll", m.handlePoll)
	http.HandleFunc("/master/results", m.handleResults)
	http.HandleFunc("/master/heartbeat", m.handleHeartbeat)
}

func (m *Master) SubmitTask(target string, plugins []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	task := &ScanTask{
		ID:      fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Target:  target,
		Plugins: plugins,
		Status:  StatusPending,
	}
	m.TaskQueue = append(m.TaskQueue, task)
}

func (m *Master) handleRegister(w http.ResponseWriter, r *http.Request) {
	var info WorkerInfo
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	m.mu.Lock()
	info.LastSeen = time.Now()
	info.Status = "active"
	m.Workers[info.ID] = &info
	m.mu.Unlock()
	log.Printf("Worker [%s] registered from %s", info.ID, info.Addr)
	w.WriteHeader(http.StatusOK)
}

func (m *Master) handlePoll(w http.ResponseWriter, r *http.Request) {
	workerID := r.URL.Query().Get("worker_id")
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.TaskQueue) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Pop task
	task := m.TaskQueue[0]
	m.TaskQueue = m.TaskQueue[1:]

	task.Status = StatusAssigned
	task.WorkerID = workerID
	m.AssignedTasks[task.ID] = task

	json.NewEncoder(w).Encode(task)
}

func (m *Master) handleResults(w http.ResponseWriter, r *http.Request) {
	var resp ScanResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return
	}

	m.mu.Lock()
	delete(m.AssignedTasks, resp.TaskID)
	m.mu.Unlock()

	log.Printf("Received results for Task [%s]", resp.TaskID)
	// Here you would publish results to the TopicRawData stream bus
}

func (m *Master) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	m.mu.Lock()
	if w, ok := m.Workers[id]; ok {
		w.LastSeen = time.Now()
		w.Status = "active"
	}
	m.mu.Unlock()
}

// startReaper detects dead workers and re-queues their tasks
func (m *Master) startReaper() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for id, worker := range m.Workers {
			if now.Sub(worker.LastSeen) > 30*time.Second {
				log.Printf("⚠️ Worker [%s] detected DEAD. Reassigning tasks...", id)
				worker.Status = "dead"
				m.reclaimTasks(id)
				delete(m.Workers, id)
			}
		}
		m.mu.Unlock()
	}
}

func (m *Master) reclaimTasks(workerID string) {
	for tid, task := range m.AssignedTasks {
		if task.WorkerID == workerID {
			if task.Retries < 3 {
				task.Status = StatusPending
				task.Retries++
				task.WorkerID = ""
				m.TaskQueue = append(m.TaskQueue, task)
				log.Printf("Re-queued Task [%s] (Retry: %d)", tid, task.Retries)
			} else {
				log.Printf("Task [%s] FAILED after max retries", tid)
			}
			delete(m.AssignedTasks, tid)
		}
	}
}
