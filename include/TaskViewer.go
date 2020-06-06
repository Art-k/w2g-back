package include

import "time"

type TaskViewer struct {
	Model
	TaskId    string
	ViewerId  string
	ExpiresAt *time.Time
}
