package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer() *Server {
	return &Server{store: NewStore()}
}

func TestHealthz(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %v", body)
	}
}

func TestCreateAndGetTask(t *testing.T) {
	srv := newTestServer()
	mux := srv.routes()

	// Create.
	createReq := httptest.NewRequest(http.MethodPost, "/tasks",
		bytes.NewBufferString(`{"title": "write tests"}`))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", createRec.Code)
	}
	var created Task
	if err := json.NewDecoder(createRec.Body).Decode(&created); err != nil {
		t.Fatalf("create: invalid JSON: %v", err)
	}
	if created.ID == 0 || created.Title != "write tests" || created.Done {
		t.Fatalf("create: unexpected task %+v", created)
	}

	// Get it back.
	getReq := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", getRec.Code)
	}
	var got Task
	if err := json.NewDecoder(getRec.Body).Decode(&got); err != nil {
		t.Fatalf("get: invalid JSON: %v", err)
	}
	if got != created {
		t.Fatalf("get: got %+v, want %+v", got, created)
	}
}

func TestListTasks(t *testing.T) {
	srv := newTestServer()
	mux := srv.routes()
	for _, title := range []string{"one", "two"} {
		req := httptest.NewRequest(http.MethodPost, "/tasks",
			bytes.NewBufferString(`{"title": "`+title+`"}`))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create %q: expected 201, got %d", title, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", rec.Code)
	}
	var tasks []Task
	if err := json.NewDecoder(rec.Body).Decode(&tasks); err != nil {
		t.Fatalf("list: invalid JSON: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("list: expected 2 tasks, got %d", len(tasks))
	}
}

func TestUpdateTask(t *testing.T) {
	srv := newTestServer()
	mux := srv.routes()
	srv.store.Create("old title")

	req := httptest.NewRequest(http.MethodPut, "/tasks/1",
		bytes.NewBufferString(`{"title": "new title", "done": true}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d", rec.Code)
	}
	var updated Task
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("update: invalid JSON: %v", err)
	}
	if updated.Title != "new title" || !updated.Done {
		t.Fatalf("update: unexpected task %+v", updated)
	}
}

func TestDeleteTask(t *testing.T) {
	srv := newTestServer()
	mux := srv.routes()
	srv.store.Create("to delete")

	delReq := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	delRec := httptest.NewRecorder()
	mux.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", delRec.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusNotFound {
		t.Fatalf("get after delete: expected 404, got %d", getRec.Code)
	}
}

func TestCreateTaskEmptyTitle(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/tasks",
		bytes.NewBufferString(`{"title": ""}`))
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetMissingTask(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/tasks/99", nil)
	rec := httptest.NewRecorder()
	srv.routes().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
