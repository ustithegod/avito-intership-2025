package user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/lib/sl"
	repo "avito-intership-2025/internal/repository"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type userService interface {
	GetReview(ctx context.Context, userID string) (*api.GetReviewResponse, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*api.UserSchema, error)
}

type UserHandler struct {
	log     *slog.Logger
	service userService
}

func NewUserHandler(log *slog.Logger, s userService) *UserHandler {
	return &UserHandler{
		log:     log,
		service: s,
	}
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"   validate:"required"`
	IsActive bool   `json:"is_active"`
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.SetIsActive"
	h.log = h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	ctx := r.Context()

	var input SetIsActiveRequest

	if err := render.DecodeJSON(r.Body, &input); err != nil {
		h.log.Error("failed to decode request body", sl.Err(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, api.Error(api.ErrBadRequest, "bad request"))
		return
	}

	if err := validator.New().Struct(input); err != nil {
		validateError := err.(validator.ValidationErrors)

		h.log.Error("invalid request", sl.Err(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, api.ValidationError(validateError))
		return
	}

	resp, err := h.service.SetIsActive(ctx, input.UserID, input.IsActive)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			h.log.Info("user not found", sl.Err(err))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, api.Error(api.ErrCodeNotFound, err.Error()))
			return
		}
		h.log.Error("error while changing user", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, api.InternalError())
		return
	}

	h.log.Info("user changed successfully")
	render.JSON(w, r, api.UserResponse{User: *resp})
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.user.GetReview"
	h.log = h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	ctx := r.Context()

	userID := r.URL.Query().Get("team_name")
	if userID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, api.Error(api.ErrBadRequest, "user_id is required"))
		return
	}

	resp, err := h.service.GetReview(ctx, userID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			h.log.Info("prs not found", sl.Err(err))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, api.Error(api.ErrCodeNotFound, err.Error()))
			return
		}
		h.log.Error("error while retrieving prs", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, api.InternalError())
		return
	}

	h.log.Info("retrieved prs successfully")
	render.JSON(w, r, resp)
}
