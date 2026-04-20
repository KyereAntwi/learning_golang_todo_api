package domain

type TodoDto struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type TodoDetailDto struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateTodoDto struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateTodoDto struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ChangeStatusDto struct {
	Status string `json:"status"`
}
