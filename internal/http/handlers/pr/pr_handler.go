package pr

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

type prService interface {
	Create(ctx context.Context, prID, prName, authorId string) (*api.PullRequestSchema, error)
	Merge(ctx context.Context, prID string) (*api.PullRequestSchema, error)
	Reassign(ctx context.Context, prID, oldRev string) (*api.ReassignResponse, error)
}

type PrHandler struct {
	log     *slog.Logger
	service prService
}

func NewPrHandler(log *slog.Logger, s prService) *PrHandler {
	return &PrHandler{
		log:     log,
		service: s,
	}
}

type CreateRequest struct {
	PrID     string `json:"pull_request_id"   validate:"required"`
	PrName   string `json:"pull_request_name" validate:"required,min=5"`
	AuthorId string `json:"author_id"         validate:"required"`
}

func (h *PrHandler) Create(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.pr.Create"
	h.log = h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	ctx := r.Context()

	var input CreateRequest
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

	resp, err := h.service.Create(ctx, input.PrID, input.PrName, input.AuthorId)
	if err != nil {
		if errors.Is(err, repo.ErrPRExists) {
			h.log.Info("pr already exists", sl.Err(err))
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, api.Error(api.ErrCodePRExists, err.Error()))
			return
		}
		if errors.Is(err, repo.ErrNotFound) {
			h.log.Info("resource not found", sl.Err(err))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, api.Error(api.ErrCodeNotFound, err.Error()))
			return
		}
		h.log.Error("error while creating pr", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, api.InternalError())
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, api.PrResponse{
		PullRequest: *resp,
	})
}

type MergeRequest struct {
	PrID string `json:"pull_request_id" validate:"required"`
}

func (h *PrHandler) Merge(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.pr.Merge"
	h.log = h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	ctx := r.Context()

	var input MergeRequest
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

	resp, err := h.service.Merge(ctx, input.PrID)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			h.log.Info("pr not found", sl.Err(err))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, api.Error(api.ErrCodeNotFound, err.Error()))
			return
		}
		h.log.Error("error while merging pr", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, api.InternalError())
		return
	}

	render.JSON(w, r, api.PrResponse{PullRequest: *resp})
}

type ReassignRequest struct {
	PrID          string `json:"pull_request_id" validate:"required"`
	OldReviewerID string `json:"old_reviewer_id" validate:"required"`
}

func (h *PrHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.pr.Merge"
	h.log = h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	ctx := r.Context()

	var input ReassignRequest
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

	resp, err := h.service.Reassign(ctx, input.PrID, input.OldReviewerID)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNotFound):
			h.log.Info("resource not found", sl.Err(err))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, api.Error(api.ErrCodeNotFound, err.Error()))

		case errors.Is(err, repo.ErrNoCandidate):
			h.log.Info("no candidate", sl.Err(err))
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, api.Error(api.ErrCodeNoCandidate, err.Error()))

		case errors.Is(err, repo.ErrPRMerged):
			h.log.Info("no candidate", sl.Err(err))
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, api.Error(api.ErrCodePRMerged, err.Error()))

		case errors.Is(err, repo.ErrNotAssigned):
			h.log.Info("no candidate", sl.Err(err))
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, api.Error(api.ErrCodeNotAssigned, err.Error()))

		default:
			h.log.Error("error while reassigning pr", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, api.InternalError())
		}
		return
	}

	render.JSON(w, r, resp)
}
