package domain

type TodoCreatedEvent struct {
	Title     string `json:"title"`
	EventType string `json:"event_type"`
}

type TodoUpdatedEvent struct {
	Title       string `json:"title"`
	ID          int64  `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	EventType   string `json:"event_type"`
}

type TodoCompletedEvent struct {
	ID        int64  `json:"id"`
	EventType string `json:"event_type"`
}
