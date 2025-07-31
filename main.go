package main

import (
	"archive-service/api"
	"archive-service/config"
	"archive-service/task"
	"fmt"
	"net/http"
	"os"
)

func main() {
	cfg := config.LoadConfig()
	taskManager := task.NewTaskManager(cfg)

	handler := api.NewHandler(taskManager, cfg)

	http.HandleFunc("/task", handler.CreateTask)

	http.HandleFunc("/task/", handler.RouteTaskSubpaths)
	http.Handle("/archives/", http.StripPrefix("/archives/", http.FileServer(http.Dir("archives"))))
	
	fmt.Println("Server listening on port:", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
