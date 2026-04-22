package board

// Board represents a named board that groups kanban tasks and calendar events.
type Board struct {
	ID        string `storm:"id" json:"id"`
	Name      string `json:"name"`
	CreatedBy string `json:"createdBy"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}
