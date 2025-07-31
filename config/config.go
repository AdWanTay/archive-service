package config

type Config struct {
	Port              string
	MaxFilesPerTask   int
	AllowedExtensions []string
	MaxActiveTasks    int
}

func LoadConfig() Config {
	return Config{
		Port:              "8080",
		MaxFilesPerTask:   3,
		AllowedExtensions: []string{".jpeg", ".pdf"},
		MaxActiveTasks:    3,
	}
}
