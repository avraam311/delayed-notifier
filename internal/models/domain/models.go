package domain

import "time"

type Notification struct {
	Message  string    `json:"message" validate:"required"`
	DateTime time.Time `json:"date_time" validate:"required"`
	Mail     string    `json:"mail" validate:"required"`
	TgID     string    `json:"tg_id"`
}

type NotificationWithID struct {
	ID       int       `json:"id"`
	Message  string    `json:"message" validate:"required"`
	DateTime time.Time `json:"date_time" validate:"required"`
	Mail     string    `json:"mail" validate:"required"`
	TgID     string    `json:"tg_id"`
}
