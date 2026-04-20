package domain

import (
	"errors"
	"time"
)

type Todo struct {
	id          int64
	title       string
	description string
	status      string
	createdAt   time.Time
	updatedAt   time.Time
}

func NewTodo(title, description string) *Todo {

	if title == "" {
		panic("title cannot be empty")
	}

	return &Todo{
		title:       title,
		description: description,
		status:      "pending",
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}
}

func (t *Todo) ChangeStatus(status string) error {

	if t.status == status {
		return errors.New("todo is already " + status)
	}

	t.status = status
	t.updatedAt = time.Now()
	return nil
}

func (t *Todo) Update(title, description string) error {
	if title == "" {
		return errors.New("title cannot be empty")
	}

	if description == "" {
		return errors.New("description cannot be empty")
	}

	t.title = title
	t.description = description
	t.updatedAt = time.Now()
	return nil
}

func (t Todo) GetID() int64 {
	return t.id
}

func (t Todo) GetTitle() string {
	return t.title
}

func (t Todo) GetDescription() string {
	return t.description
}

func (t Todo) GetStatus() string {
	return t.status
}

func (t Todo) GetCreatedAt() time.Time {
	return t.createdAt
}

func (t Todo) GetUpdatedAt() time.Time {
	return t.updatedAt
}

func (t *Todo) SetID(id int64) {
	t.id = id
}

func (t *Todo) SetCreatedAt(createdAt time.Time) {
	t.createdAt = createdAt
}

func (t *Todo) SetUpdatedAt(updatedAt time.Time) {
	t.updatedAt = updatedAt
}
