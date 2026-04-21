package repositories

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"todoapi.com/m/src/domain"
	"todoapi.com/m/src/utils"
)

type TodoRepository struct {
	db         *sql.DB
	rabbitConn *utils.RabbitMQConnection
}

func NewTodoRepository(db *sql.DB, rabbitConn *utils.RabbitMQConnection) *TodoRepository {
	return &TodoRepository{db: db, rabbitConn: rabbitConn}
}

func (r *TodoRepository) Create(title, description string) (int64, error) {
	todo := domain.NewTodo(title, description)

	query := `
	INSERT INTO todos (title, description, status, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING id`

	args := []interface{}{todo.GetTitle(), todo.GetDescription(), todo.GetStatus(), todo.GetCreatedAt(), todo.GetUpdatedAt()}

	var id int64
	err := r.db.QueryRow(query, args...).Scan(&id)

	if err != nil {
		return 0, err
	}

	todo.SetID(id)

	return todo.GetID(), nil
}

func (r *TodoRepository) GetTodoByID(id int64) (domain.Todo, error) {
	if id <= 0 {
		return domain.Todo{}, errors.New("invalid id")
	}

	query := `
	SELECT id, title, description, status, created_at, updated_at 
	FROM todos 
	WHERE id = $1`

	var title, description, status string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(query, id).Scan(&id, &title, &description, &status, &createdAt, &updatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return domain.Todo{}, errors.New("todo not found")
		default:
			return domain.Todo{}, err
		}
	}

	todo := domain.NewTodo(title, description)
	todo.SetID(id)
	todo.SetCreatedAt(createdAt)
	todo.SetUpdatedAt(updatedAt)
	todo.ChangeStatus(status)

	return *todo, nil
}

func (r *TodoRepository) Update(id int64, title, description string) error {
	if id <= 0 {
		return errors.New("invalid id")
	}

	todo, err := r.GetTodoByID(id)
	if err != nil {
		return err
	}

	// log the current state of the todo before updating
	log.Printf("Updating todo ID %d: current title=%s, description=%s, status=%s", id, todo.GetTitle(), todo.GetDescription(), todo.GetStatus())
	// publish an event before updating the todo
	var publisherService IPublisherServices = NewPublisherService(r.rabbitConn)
	err = publisherService.PublishTodoUpdated(domain.TodoUpdatedEvent{
		Title:       todo.GetTitle(),
		ID:          todo.GetID(),
		Description: todo.GetDescription(),
		Status:      todo.GetStatus(),
		CreatedAt:   todo.GetCreatedAt().Format(time.RFC3339),
		UpdatedAt:   todo.GetUpdatedAt().Format(time.RFC3339),
		EventType:   "todo_updated",
	})
	if err != nil {
		log.Printf("Error publishing todo updated event: %v", err)
		// we can choose to return the error or ignore it based on the requirements
		// for now, we'll just log it and continue with the update
	}

	err = todo.Update(title, description)
	if err != nil {
		return err
	}

	query := `
	UPDATE todos 
	SET title = $1, description = $2, updated_at = $3 
	WHERE id = $4`

	args := []interface{}{todo.GetTitle(), todo.GetDescription(), todo.GetUpdatedAt(), id}

	_, err = r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *TodoRepository) Delete(id int64) error {
	if id <= 0 {
		return errors.New("invalid id")
	}

	query := `
	DELETE FROM todos 
	WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *TodoRepository) ChangeStatus(id int64, status string) error {
	if id <= 0 {
		return errors.New("invalid id")
	}

	todo, err := r.GetTodoByID(id)
	if err != nil {
		return err
	}

	err = todo.ChangeStatus(status)
	if err != nil {
		return err
	}

	query := `
	UPDATE todos 
	SET status = $1, updated_at = $2 
	WHERE id = $3`

	args := []interface{}{todo.GetStatus(), todo.GetUpdatedAt(), id}

	_, err = r.db.Exec(query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *TodoRepository) GetAll(searchKey string, page int64, pageSize int64) ([]domain.Todo, error) {
	if r.db == nil {
		return []domain.Todo{}, errors.New("database connection is nil")
	}

	offset := (page - 1) * pageSize

	query := `
	SELECT id, title, description, status
	FROM todos 
	WHERE title LIKE $1 OR description LIKE $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, "%"+searchKey+"%", pageSize, offset)
	if err != nil {
		return []domain.Todo{}, err
	}
	defer rows.Close()

	var todos []domain.Todo

	for rows.Next() {
		var id int64
		var title, description, status string

		err := rows.Scan(&id, &title, &description, &status)
		if err != nil {
			return []domain.Todo{}, err
		}

		todo := domain.NewTodo(title, description)
		todo.SetID(id)
		todo.ChangeStatus(status)

		todos = append(todos, *todo)
	}

	if err = rows.Err(); err != nil {
		return []domain.Todo{}, err
	}

	return todos, nil
}

func (r *TodoRepository) DoesTodoExist(title string) (bool, error) {
	if title == "" {
		return false, errors.New("title cannot be empty")
	}

	query := `
	SELECT id
	FROM todos 
	WHERE title = $1`

	var id int64
	err := r.db.QueryRow(query, title).Scan(&id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
