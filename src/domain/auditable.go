package domain

import "time"

type Auditable struct {
	Id         int64
	EntityType string
	EntityId   int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Data       string
}
