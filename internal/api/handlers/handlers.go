package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/avraam311/delayed-notifier/internal/models/domain"

	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

type notificationService interface {
	CreateNotification(context.Context, *domain.Notification) (int, error)
	GetNotificationStatus(context.Context, int) (string, error)
	DeleteNotification(context.Context, int) error
}

type HandlerNotification struct {
	service   notificationService
	validator *validator.Validate
}

func NewHandler(s notificationService, v *validator.Validate) *HandlerNotification {
	return &HandlerNotification{
		service:   s,
		validator: v,
	}
}

func (h *HandlerNotification) CreateNotification(c *ginext.Context) {
	var not *domain.Notification

	if err := json.NewDecoder(c.Request.Body).Decode(&not); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to decode request body")
		Fail(c.Writer, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validator.Struct(not); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to validate request body")
		Fail(c.Writer, http.StatusBadRequest, "invalid json - "+err.Error())
		return
	}

	id, err := h.service.CreateNotification(c.Request.Context(), not)
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to create notification")
		Fail(c.Writer, http.StatusInternalServerError, "internal error")
		return
	}

	Created(c.Writer, id)
}

func (h *HandlerNotification) GetNotificationStatus(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("id is not proper integer")
		Fail(c.Writer, http.StatusBadRequest, "invalid id integer")
		return
	}

	if id < 1 {
		zlog.Logger.Warn().Err(err).Msg("negative or zero id")
		Fail(c.Writer, http.StatusBadRequest, "id must be > 0")
		return
	}

	status, err := h.service.GetNotificationStatus(c.Request.Context(), id)
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to get notification status")
		Fail(c.Writer, http.StatusInternalServerError, "internal error")
		return
	}

	OK(c.Writer, status)
}

func (h *HandlerNotification) DeleteNotification(c *ginext.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		zlog.Logger.Warn().Err(err).Msg("id is not proper integer")
		Fail(c.Writer, http.StatusBadRequest, "invalid id integer")
		return
	}

	if id < 1 {
		zlog.Logger.Warn().Err(err).Msg("negative or zero id")
		Fail(c.Writer, http.StatusBadRequest, "id must be > 0")
		return
	}

	if err := h.service.DeleteNotification(c.Request.Context(), id); err != nil {
		zlog.Logger.Warn().Err(err).Msg("failed to delete notification")
		Fail(c.Writer, http.StatusInternalServerError, "internal error")
		return
	}

	OK(c.Writer, "notification deleted")
}
