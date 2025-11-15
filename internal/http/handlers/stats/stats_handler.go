package stats

import (
	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/lib/sl"
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type statsService interface {
	GetStatistics(ctx context.Context, sort string) (*api.StatsResponse, error)
}

type StatsHandler struct {
	log     *slog.Logger
	service statsService
}

func NewStatsHandler(log *slog.Logger, s statsService) *StatsHandler {
	return &StatsHandler{
		log:     log,
		service: s,
	}
}

func (h *StatsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.stats.GetStatistics"
	log := h.log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	ctx := r.Context()

	sort := r.URL.Query().Get("sort")
	sort = strings.ToLower(sort)
	if sort == "" {
		sort = "desc"
	} else if sort != "desc" && sort != "asc" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, api.Error(api.ErrBadRequest, "sort must be 'desc' or 'asc'. can be omitted: 'desc' by default"))
		return
	}

	resp, err := h.service.GetStatistics(ctx, sort)
	if err != nil {
		log.Error("error while retrieving statistics", sl.Err(err))
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, api.InternalError())
	}
	render.JSON(w, r, resp)
}
