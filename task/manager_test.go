package task

import (
	"archive-service/config"
	"archive-service/utils"
	"errors"
	"testing"
	"time"
)

func TestCreateTask(t *testing.T) {
	cfg := config.Config{
		MaxActiveTasks:    2,
		MaxFilesPerTask:   3,
		AllowedExtensions: []string{".pdf", ".jpeg"},
		Port:              "8080",
	}

	m := NewTaskManager(cfg)

	id1, err := m.CreateTask()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id1 == "" {
		t.Fatalf("Expected non-empty task id")
	}

	id2, err := m.CreateTask()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if id2 == "" {
		t.Fatalf("Expected non-empty task id")
	}

	_, err = m.CreateTask()
	if err == nil {
		t.Fatalf("Expected error on exceeding max active tasks")
	}
	if err.Error() != "too many active tasks" {
		t.Fatalf("Expected error 'too many active tasks', got %v", err)
	}
}

func TestGetStatus(t *testing.T) {

	cfg := config.Config{
		MaxActiveTasks:    3,
		MaxFilesPerTask:   3,
		AllowedExtensions: []string{".pdf", ".jpeg"},
		Port:              "8080",
	}

	manager := NewTaskManager(cfg)

	_, err := manager.GetStatus("invalid_id")
	if err == nil {
		t.Fatal("Expected error for non-existent task ID, got nil")
	}
	if err.Error() != "task not found" {
		t.Fatalf("Expected 'task not found', got: %v", err)
	}

	taskID, err := manager.CreateTask()
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	manager.mu.Lock()
	manager.tasks[taskID].Status = Done
	manager.tasks[taskID].BadLinks = []string{"http://askd/file.jpeg"}
	manager.tasks[taskID].ArchiveURL = "/archives/" + taskID + ".zip"
	manager.mu.Unlock()

	status, err := manager.GetStatus(taskID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if status["status"] != Done {
		t.Errorf("Expected status 'Done', got: %v", status["status"])
	}
	if url, ok := status["archive_url"].(string); !ok || url == "" {
		t.Errorf("Expected non-empty archive_url, got: %v", status["archive_url"])
	}
	if errors, ok := status["error_files"].([]string); !ok || len(errors) != 1 {
		t.Errorf("Expected 1 bad link, got: %v", status["error_files"])
	}
}

func TestAddFileAndProcessTask(t *testing.T) {
	origDownload := utils.DownloadAndZip
	defer func() { utils.DownloadAndZip = origDownload }()

	utils.DownloadAndZip = func(id string, urls []string) (string, []string, error) {
		return "archives/" + id + ".zip", []string{}, nil
	}

	cfg := config.Config{
		MaxActiveTasks:    1,
		MaxFilesPerTask:   3,
		AllowedExtensions: []string{".pdf", ".jpeg"},
		Port:              "8080",
	}

	tm := NewTaskManager(cfg)

	taskID, err := tm.CreateTask()
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}

	err = tm.AddFile(taskID, "http://example.com/file1.pdf")
	if err != nil {
		t.Fatalf("AddFile 1 failed: %v", err)
	}
	err = tm.AddFile(taskID, "http://example.com/file2.jpeg")
	if err != nil {
		t.Fatalf("AddFile 2 failed: %v", err)
	}
	err = tm.AddFile(taskID, "http://example.com/file3.pdf")
	if err != nil {
		t.Fatalf("AddFile 3 failed: %v", err)
	}

	err = tm.AddFile(taskID, "http://example.com/file4.pdf")
	if err == nil || err.Error() != "task file limit reached" {
		t.Errorf("Expected file limit error, got: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	status, err := tm.GetStatus(taskID)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status["status"] != Done {
		t.Errorf("Expected status Done, got: %v", status["status"])
	}

	if status["archive_url"] == "" {
		t.Errorf("Expected non-empty archive_url")
	}
}

func TestProcessTask_Error(t *testing.T) {
	origDownload := utils.DownloadAndZip
	defer func() { utils.DownloadAndZip = origDownload }()
	utils.DownloadAndZip = func(id string, urls []string) (string, []string, error) {
		return "", []string{"http://badfile.com/file.pdf"}, errors.New("download error")
	}

	cfg := config.Config{
		MaxActiveTasks:    1,
		MaxFilesPerTask:   3,
		AllowedExtensions: []string{".pdf", ".jpeg"},
		Port:              "8080",
	}

	tm := NewTaskManager(cfg)
	taskID, _ := tm.CreateTask()

	tm.AddFile(taskID, "http://sadad/file1.pdf")
	tm.AddFile(taskID, "http://sadad/file2.jpeg")
	tm.AddFile(taskID, "http://sadad/file3.pdf")

	time.Sleep(200 * time.Millisecond)

	status, err := tm.GetStatus(taskID)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status["status"] != Error {
		t.Errorf("Expected status Error, got: %v", status["status"])
	}
	if files, ok := status["error_files"].([]string); !ok || len(files) == 0 {
		t.Errorf("Expected error_files to be non-empty, got: %v", status["error_files"])
	}
}
