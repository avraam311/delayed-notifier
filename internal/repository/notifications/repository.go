package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/avraam311/delayed-notifier/internal/models/domain"

	"github.com/wb-go/wbf/dbpg"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

const (
	notStatus = "in queue"
)

type repositoryNotification struct {
	db *dbpg.DB
}

func NewRepository(db *dbpg.DB) *repositoryNotification {
	return &repositoryNotification{
		db: db,
	}
}

func (r *repositoryNotification) CreateNotification(ctx context.Context, not *domain.Notification) (int, error) {
	query := `
		INSERT INTO notification (
			message, date_time, tg_id, mail, status
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	var id int
	err := r.db.QueryRowContext(
		ctx, query, not.Message, not.DateTime, not.TgID, not.Mail, notStatus,
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
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotificationNotFound
		}

		return "", fmt.Errorf("notifications/repository.go - %w", err)
	}

	return status, nil
}

func (r *repositoryNotification) DeleteNotification(ctx context.Context, id int) error {
	query := `
		INSERT INTO delete (delete_id)
		SELECT $1
		WHERE EXISTS (
			SELECT 1 FROM notification WHERE id = $1
		)
		ON CONFLICT (delete_id) DO NOTHING;
	`

	_, err := r.db.ExecContext(
		ctx, query, id,
	)
	if err != nil {
		return fmt.Errorf("notifications/repository.go - %w", err)
	}

	return nil
}

func (r *repositoryNotification) CheckIfToDelete(ctx context.Context, deleteID int) (int, error) {
	query := `
		SELECT id
		FROM delete
		WHERE delete_id = $1
	`

	var id int
	err := r.db.QueryRowContext(
		ctx, query, deleteID,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, fmt.Errorf("notifications/repository.go - %w", err)
	}

	return id, nil
}

func (r *repositoryNotification) ChangeNotificationStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE notification
		SET status = $2
		WHERE id =  $1
	`

	res, err := r.db.ExecContext(
		ctx, query, id, status,
	)
	if err != nil {
		return fmt.Errorf("notifications/repository.go - %w", err)
	}

	affectedRows, _ := res.RowsAffected()
	if affectedRows == 0 {
		return ErrNotificationNotFound
	}

	return nil
}
