package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/avraam311/delayed-notifier/internal/models/domain"
)

type repositoryNotification interface {
	CreateNotification(context.Context, *domain.Notification) (int, error)
	GetNotificationStatus(context.Context, int) (string, error)
	DeleteNotification(context.Context, int) error
}

type rabbitMQ interface {
	Publish(string, []byte, string, time.Time) error
}

type ServiceNotification struct {
	repo     repositoryNotification
	rabbitMQ rabbitMQ
}

func NewService(repo repositoryNotification) *ServiceNotification {
	return &ServiceNotification{
		repo: repo,
	}
}

func (s *ServiceNotification) CreateNotification(ctx context.Context, not *domain.Notification) (int, error) {
	id, err := s.repo.CreateNotification(ctx, not)
	if err != nil {
		return 0, err
	}

	delay := not.DateTime
	msg, err := json.Marshal(not)
	if err != nil {
		return 0, fmt.Errorf("notification/service.go - %w", err)
	}

	s.rabbitMQ.Publish("notifications-key", msg, "application/json", delay)

	return id, nil
}

func (s *ServiceNotification) GetNotificationStatus(ctx context.Context, id int) (string, error) {
	status, err := s.repo.GetNotificationStatus(ctx, id)
	if err != nil {
		return "", err
	}

	return status, nil
}

func (s *ServiceNotification) DeleteNotification(ctx context.Context, id int) error {
	err := s.repo.DeleteNotification(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
