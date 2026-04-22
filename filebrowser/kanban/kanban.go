package kanban

// Task represents a kanban board task.
type Task struct {
	ID          uint   `storm:"id,increment" json:"id"`
	BoardID     string `storm:"index" json:"boardId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Column      string `storm:"index" json:"column"`
	Position    int    `json:"position"`
	AssignedTo  string `json:"assignedTo"`
	StartDate   string `json:"startDate,omitempty"`
	EndDate     string `json:"endDate,omitempty"`
	CreatedBy   string `json:"createdBy"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}
