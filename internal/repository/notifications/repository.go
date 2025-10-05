package notifications

import (
	"context"
	"fmt"

	"github.com/avraam311/delayed-notifier/internal/models/domain"

	"github.com/wb-go/wbf/dbpg"
)

type repositoryNotification struct {
	db *dbpg.DB
}

func NewRepositoryNotification(db *dbpg.DB) *repositoryNotification {
	return &repositoryNotification{
		db: db,
	}
}

func (r *repositoryNotification) CreateNotification(ctx context.Context, not *domain.Notification) (int, error) {
	query := `
		INSERT INTO notification (
			message, date_time, tg_id, mail
		) VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	var id int
	err := r.db.QueryRowContext(
		ctx, query, not.Message, not.DateTime, not.TgID, not.Mail,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("notifications/repository.go - %w", err)
	}

	return id, nil
}

func (r *repositoryNotification) GetNotificationStatus(ctx context.Context, id int) (string, error) {
	query := `
		SELECT status
		FROM notification
		WHERE id = $1
	`

	var status string
	err := r.db.QueryRowContext(
		ctx, query, id,
	).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("notifications/repository.go - %w", err)
	}

	return status, nil
}

func (r *repositoryNotification) DeleteNotification(ctx context.Context, id int) error {
	query := `
		DELETE
		FROM notifications
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx, query, id,
	)
	if err != nil {
		return fmt.Errorf("notifications/repository.go - %w", err)
	}

	return nil
}
