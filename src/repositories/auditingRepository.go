package repositories

import (
	"database/sql"
	"time"

	"todoapi.com/m/src/domain"
)

type AuditingRepository struct {
	db *sql.DB
}

func NewAuditingRepository(db *sql.DB) *AuditingRepository {
	return &AuditingRepository{db: db}
}

func (r *AuditingRepository) RecordAuditEntry(entityType string, entityId int64, data string) error {
	var auditableEntry = domain.Auditable{
		EntityType: entityType,
		EntityId:   entityId,
		Data:       data,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	query := `
	INSERT INTO audit_logs (entity_type, entity_id, data, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(query, auditableEntry.EntityType, auditableEntry.EntityId, auditableEntry.Data, auditableEntry.CreatedAt, auditableEntry.UpdatedAt)
	return err
}
