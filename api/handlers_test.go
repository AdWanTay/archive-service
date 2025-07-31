package api

import (
	"archive-service/config"
	"archive-service/task"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testServer struct {
	handler *Handler
}

func setupTestServer() *testServer {
	cfg := config.LoadConfig()
	tm := task.NewTaskManager(cfg)
	h := NewHandler(tm, cfg)
	return &testServer{handler: h}
}

func TestCreateTask_Success(t *testing.T) {
	ts := setupTestServer()
	req := httptest.NewRequest(http.MethodPost, "/task", nil)
	w := httptest.NewRecorder()

	ts.handler.CreateTask(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", res.StatusCode)
	}
}

func TestCreateTask_MethodNotAllowed(t *testing.T) {
	ts := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/task", nil)
	w := httptest.NewRecorder()

	ts.handler.CreateTask(w, req)
	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", w.Result().StatusCode)
	}
}

func TestRouteTaskSubpaths_AddFile(t *testing.T) {
	ts := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/task", nil)
	w := httptest.NewRecorder()
	ts.handler.CreateTask(w, req)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	taskID := resp["task_id"]

	body := bytes.NewBufferString(`{"url":"http://example.com/file.pdf"}`)
	req = httptest.NewRequest(http.MethodPost, "/task/"+taskID+"/add", body)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	ts.handler.RouteTaskSubpaths(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200 on AddFile, got %d", w.Result().StatusCode)
	}
}

func TestRouteTaskSubpaths_Status(t *testing.T) {
	ts := setupTestServer()

	req := httptest.NewRequest(http.MethodPost, "/task", nil)
	w := httptest.NewRecorder()
	ts.handler.CreateTask(w, req)

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	taskID := resp["task_id"]

	req = httptest.NewRequest(http.MethodGet, "/task/"+taskID+"/status", nil)
	w = httptest.NewRecorder()
	ts.handler.RouteTaskSubpaths(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200 on status, got %d", w.Result().StatusCode)
	}
}

func TestAddFile_InvalidJSON(t *testing.T) {
	ts := setupTestServer()
	body := strings.NewReader("not a json")
	req := httptest.NewRequest(http.MethodPost, "/task/123/add", body)
	w := httptest.NewRecorder()
	ts.handler.AddFile(w, req, "123")
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Result().StatusCode)
	}
}

func TestAddFile_EmptyURL(t *testing.T) {
	ts := setupTestServer()
	body := strings.NewReader(`{"url":""}`)
	req := httptest.NewRequest(http.MethodPost, "/task/123/add", body)
	w := httptest.NewRecorder()
	ts.handler.AddFile(w, req, "123")
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty URL, got %d", w.Result().StatusCode)
	}
}

func TestStatus_NotFound(t *testing.T) {
	ts := setupTestServer()
	w := httptest.NewRecorder()
	ts.handler.Status(w, "unknown")
	if w.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Result().StatusCode)
	}
}

func TestStatus_Success(t *testing.T) {
	ts := setupTestServer()

	taskID, err := ts.handler.tm.CreateTask()
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	w := httptest.NewRecorder()
	ts.handler.Status(w, taskID)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if _, ok := data["status"]; !ok {
		t.Error("Response JSON missing 'status' field")
	}
	if _, ok := data["archive_url"]; !ok {
		t.Error("Response JSON missing 'archive_url' field")
	}
}
