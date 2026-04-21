package main

import (
	"database/sql"
	"encoding/json"
	"log"

	"todoapi.com/m/src/domain"
	"todoapi.com/m/src/repositories"
)

func persistUpdatedTodoAuditEntry(messageBody []byte, db *sql.DB) error {
	log.Println("Persisting audit entry for todo_updated event")
	var todoUpdatedEvent domain.TodoUpdatedEvent

	err := json.Unmarshal(messageBody, &todoUpdatedEvent)

	if err != nil {
		log.Printf("Error unmarshaling todo_updated event: %v", err)
		return err
	}

	var auditingRepo repositories.IAuditingRepository = repositories.NewAuditingRepository(db)

	err = auditingRepo.RecordAuditEntry("TodoUpdatedEvent", todoUpdatedEvent.ID, string(messageBody))
	if err != nil {
		log.Printf("Error recording audit entry: %v", err)
		return err
	} else {
		log.Printf("Successfully recorded audit entry for todo ID %d", todoUpdatedEvent.ID)
	}
	return nil
}
