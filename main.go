// Command go-microservice-cicd is a small REST API for managing tasks.
//
// It is a portfolio demonstration project: tasks live in memory, there is no
// database, and it is not a production service. It exists to show Go, HTTP
// handlers, unit testing, container builds, and CI/CD in one place.
//
// Endpoints:
//
//	GET    /healthz       - liveness/readiness probe
//	GET    /tasks         - list all tasks
//	POST   /tasks         - create a task {"title": "..."}
//	GET    /tasks/{id}    - get one task
//	PUT    /tasks/{id}    - update a task {"title": "...", "done": true}
//	DELETE /tasks/{id}    - delete a task
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// Task is a single to-do item.
type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Store is a thread-safe in-memory task store.
type Store struct {
	mu     sync.Mutex
	tasks  map[int]Task
	nextID int
}

// NewStore returns an empty Store.
func NewStore() *Store {
	return &Store{tasks: make(map[int]Task), nextID: 1}
}

func (s *Store) Create(title string) Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Task{ID: s.nextID, Title: title}
	s.nextID++
	s.tasks[t.ID] = t
	return t
}

func (s *Store) Get(id int) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *Store) List() []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	return out
}

func (s *Store) Update(id int, title string, done bool) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, false
	}
	t.Title = title
	t.Done = done
	s.tasks[id] = t
	return t, true
}

func (s *Store) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return false
	}
	delete(s.tasks, id)
	return true
}

// Server wires the store to HTTP handlers.
type Server struct {
	store *Store
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.List())
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	writeJSON(w, http.StatusCreated, s.store.Create(body.Title))
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}
	t, ok := s.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}
	var body struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	t, ok := s.store.Update(id, body.Title, body.Done)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}
	if !s.store.Delete(id) {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// routes returns the HTTP mux for the server.
func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /tasks", s.handleListTasks)
	mux.HandleFunc("POST /tasks", s.handleCreateTask)
	mux.HandleFunc("GET /tasks/{id}", s.handleGetTask)
	mux.HandleFunc("PUT /tasks/{id}", s.handleUpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", s.handleDeleteTask)
	return mux
}

func main() {
	srv := &Server{store: NewStore()}
	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, srv.routes()))
}
