package event

// Event represents a calendar event.
type Event struct {
	ID        uint   `storm:"id,increment" json:"id"`
	BoardID   string `storm:"index" json:"boardId"`
	Title     string `json:"title"`
	Date      string `json:"date"`
	EndDate   string `json:"endDate,omitempty"`
	Color     string `json:"color,omitempty"`
	CreatedBy string `json:"createdBy"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}
