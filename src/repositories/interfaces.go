package repositories

import (
	"github.com/google/uuid"
	"todoapi.com/m/src/domain"
)

type ITodoRepository interface {
	Create(title, description string, createdBy uuid.UUID) (int64, error)
	GetTodoByID(id int64) (domain.Todo, error)
	Update(id int64, title, description string) error
	Delete(id int64) error
	ChangeStatus(id int64, status string) error
	GetAll(searchKey string, createdBy uuid.UUID, page int64, pageSize int64) ([]domain.Todo, error)
	DoesTodoExist(title string) (bool, error)
}

type IUserRepository interface {
	Create(username, password, email, primaryPhone string) (uuid.UUID, error)
	GetUserByID(id uuid.UUID) (domain.User, error)
	GetUserByUsername(username string) (domain.User, error)
	IsUsernameTaken(username string) (bool, error)
	Authenticate(username, password string) (string, error)
	StoreRefreshToken(token domain.RefreshToken) error
	GetRefreshTokenByUserID(userID uuid.UUID) (domain.RefreshToken, error)
	DeleteRefreshToken(userID uuid.UUID) error
}

type IPublisherServices interface {
	PublishTodoCreated(todo domain.TodoCreatedEvent) error
	PublishTodoUpdated(todo domain.TodoUpdatedEvent) error
	PublishTodoCompleted(todo domain.TodoCompletedEvent) error
}

type IAuditingRepository interface {
	RecordAuditEntry(entityType string, entityId int64, data string) error
}
