package task

import (
	"archive-service/config"
	"archive-service/utils"
	"errors"
	"fmt"
	"sync"
)

type Manager struct {
	cfg     config.Config
	mu      sync.Mutex
	tasks   map[string]*Task
	working int
}

func NewTaskManager(cfg config.Config) *Manager {
	return &Manager{
		tasks: make(map[string]*Task),
		cfg:   cfg,
	}
}

func (tm *Manager) CreateTask() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.working >= tm.cfg.MaxActiveTasks {
		return "", errors.New("too many active tasks")
	}
	id := utils.GenerateID()
	tm.tasks[id] = &Task{ID: id, Status: Pending}
	tm.working++
	return id, nil
}

func (tm *Manager) AddFile(taskID, url string) error {
	tm.mu.Lock()
	task, ok := tm.tasks[taskID]
	tm.mu.Unlock()
	if !ok {
		return errors.New("task not found")
	}
	if len(task.Files) >= tm.cfg.MaxFilesPerTask {
		return errors.New("task file limit reached")
	}
	if !utils.ValidExtension(url, tm.cfg.AllowedExtensions) {
		return errors.New("unsupported file extension")
	}
	task.Files = append(task.Files, url)

	if len(task.Files) == tm.cfg.MaxFilesPerTask {
		go tm.processTask(task)
	}
	return nil
}

func (tm *Manager) GetStatus(taskID string) (map[string]interface{}, error) {
	tm.mu.Lock()
	task, ok := tm.tasks[taskID]
	tm.mu.Unlock()
	if !ok {
		return nil, errors.New("task not found")
	}
	return map[string]interface{}{
		"status":      task.Status,
		"error_files": task.BadLinks,
		"archive_url": task.ArchiveURL,
	}, nil
}

func (tm *Manager) processTask(t *Task) {
	tm.mu.Lock()
	t.Status = Progress
	tm.mu.Unlock()

	zipPath, badLinks, err := utils.DownloadAndZip(t.ID, t.Files)
	t.BadLinks = badLinks

	if err != nil || len(zipPath) == 0 {
		t.Status = Error
	} else {
		t.ArchiveURL = fmt.Sprintf("http://localhost:8080/%s", zipPath)
		t.Status = Done
	}

	tm.mu.Lock()
	tm.working--
	tm.mu.Unlock()
}
