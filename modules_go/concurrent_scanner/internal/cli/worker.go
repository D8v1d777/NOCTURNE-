package distribution

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"nocturne/scanner/internal/engine"
	"nocturne/scanner/internal/models"
	"time"
)

type Worker struct {
	ID         string
	MasterAddr string
	Manager    *engine.Manager
}

func NewWorker(id string, master string, manager *engine.Manager) *Worker {
	return &Worker{ID: id, MasterAddr: master, Manager: manager}
}

func (w *Worker) Start() {
	log.Printf("Worker [%s] connecting to Master at %s", w.ID, w.MasterAddr)

	// 1. Register
	w.register()

	// 2. Heartbeat loop
	go w.heartbeat()

	// 3. Task poll loop
	for {
		w.pollAndExecute()
		time.Sleep(2 * time.Second)
	}
}

func (w *Worker) register() {
	info := WorkerInfo{ID: w.ID, Addr: "localhost"}
	body, _ := json.Marshal(info)
	http.Post(w.MasterAddr+"/master/register", "application/json", bytes.NewBuffer(body))
}

func (w *Worker) heartbeat() {
	for {
		http.Post(w.MasterAddr+"/master/heartbeat?id="+w.ID, "", nil)
		time.Sleep(5 * time.Second)
	}
}

func (w *Worker) pollAndExecute() {
	resp, err := http.Get(w.MasterAddr + "/master/tasks/poll?worker_id=" + w.ID)
	if err != nil || resp.StatusCode != http.StatusOK {
		return
	}
	defer resp.Body.Close()

	var task ScanTask
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return
	}

	log.Printf("🚀 Executing Task [%s] for Target: %s", task.ID, task.Target)

	results := w.Manager.RunPlugins(task.Target, task.Plugins)

	w.reportResults(task.ID, results)
}

func (w *Worker) reportResults(taskID string, results interface{}) {
	payload := ScanResponse{
		TaskID:  taskID,
		Results: results.([]models.Result), // Simplified cast
	}
	body, _ := json.Marshal(payload)
	http.Post(w.MasterAddr+"/master/results", "application/json", bytes.NewBuffer(body))
	log.Printf("✅ Task [%s] completed", taskID)
}
