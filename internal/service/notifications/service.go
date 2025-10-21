package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/avraam311/delayed-notifier/internal/models/domain"
	"github.com/wb-go/wbf/config"
)

type repositoryNotification interface {
	CreateNotification(context.Context, *domain.Notification) (int, error)
	GetNotificationStatus(context.Context, int) (string, error)
	DeleteNotification(context.Context, int) error
}

type rabbitMQ interface {
	Publish(string, []byte, string, time.Duration) error
}

type ServiceNotification struct {
	repo     repositoryNotification
	rabbitMQ rabbitMQ
	cfg      *config.Config
}

func NewService(repo repositoryNotification, rMQ rabbitMQ, cfg *config.Config) *ServiceNotification {
	return &ServiceNotification{
		repo:     repo,
		rabbitMQ: rMQ,
		cfg:      cfg,
	}
}

func (s *ServiceNotification) CreateNotification(ctx context.Context, not *domain.Notification) (int, error) {
	id, err := s.repo.CreateNotification(ctx, not)
	if err != nil {
		return 0, err
	}

	notToPublish := &domain.NotificationWithID{
		ID:       id,
		Message:  not.Message,
		DateTime: not.DateTime,
		Mail:     not.Mail,
		TgID:     not.TgID,
	}

	msg, err := json.Marshal(notToPublish)
	if err != nil {
		return 0, fmt.Errorf("notification/service.go - failed to marshal notification into json - %w", err)
	}

	delay := time.Until(notToPublish.DateTime)
	err = s.rabbitMQ.Publish(s.cfg.GetString("rabbitmq.routing_key"), msg, "application/json", delay)
	if err != nil {
		return 0, fmt.Errorf("notification/service.go - failed to publish message into rabbitmq %w", err)
	}

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
