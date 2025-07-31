package task

type Status string

const (
	Pending  Status = "pending"
	Progress Status = "in_progress"
	Done     Status = "done"
	Error    Status = "error"
)

type Task struct {
	ID         string
	Files      []string
	BadLinks   []string
	ArchiveURL string
	Status     Status
}
