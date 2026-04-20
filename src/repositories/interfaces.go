package repositories

import "todoapi.com/m/src/domain"

type ITodoRepository interface {
	Create(title, description string) (int64, error)
	GetTodoByID(id int64) (domain.Todo, error)
	Update(id int64, title, description string) error
	Delete(id int64) error
	ChangeStatus(id int64, status string) error
	GetAll(searchKey string, page int64, pageSize int64) ([]domain.Todo, error)
	DoesTodoExist(title string) (bool, error)
}
