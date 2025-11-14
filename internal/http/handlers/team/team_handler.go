package team

import (
	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/lib/sl"
	repo "avito-intership-2025/internal/repository"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type teamService interface {
	Add(ctx context.Context, teamName string, users []api.TeamMember) (*api.TeamSchema, error)
	Get(ctx context.Context, teamName string) (*api.TeamSchema, error)
}

type TeamHandler struct {
	log     *slog.Logger
	service teamService
}

func NewTeamHandler(log *slog.Logger, s teamService) *TeamHandler {
	return &TeamHandler{
		log:     log,
		service: s,
	}
}

type TeamAddRequest struct {
	TeamName string           `json:"team_name" validate:"required,max=16"`
	Members  []api.TeamMember `json:"members" validate:"required,dive"`
}

func (h *TeamHandler) Add(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.team.Add"
	h.log = h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	ctx := r.Context()

	var input TeamAddRequest

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

	resp, err := h.service.Add(ctx, input.TeamName, input.Members)
	if err != nil {
		if errors.Is(err, repo.ErrTeamExists) {
			h.log.Error("team exists", sl.Err(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, api.Error(api.ErrCodeTeamExists, err.Error()))
			return
		}
		h.log.Error("error while saving team", sl.Err(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, api.InternalError())
		return
	}

	h.log.Info("team created successfully")
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, api.TeamResponse{Team: *resp})
}

func (h *TeamHandler) Get(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.team.Get"
	h.log = h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	ctx := r.Context()

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, api.Error(api.ErrBadRequest, "team_name is required"))
		return
	}

	resp, err := h.service.Get(ctx, teamName)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			h.log.Info("team not found", sl.Err(err))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, api.Error(api.ErrCodeNotFound, err.Error()))
			return
		}
		h.log.Error("error while retrieving team", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, api.InternalError())
		return
	}
	h.log.Info("team retrieved")
	render.JSON(w, r, resp)
}
