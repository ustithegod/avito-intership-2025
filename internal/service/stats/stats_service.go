package stats

import (
	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/models"
	"avito-intership-2025/internal/service"
	"context"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=StatsProvider
type StatsProvider interface {
	GetAssignmentsCountStats(ctx context.Context, sort string) ([]*models.UserStatistics, error)
	GetPrStats(ctx context.Context) (*models.PrStatistics, error)
}

type StatsService struct {
	statsProvider StatsProvider
	trm           service.TransactionManager
}

func NewStatsService(trm service.TransactionManager, statsProvider StatsProvider) *StatsService {
	return &StatsService{
		trm:           trm,
		statsProvider: statsProvider,
	}
}

func (s *StatsService) GetStatistics(ctx context.Context, sort string) (*api.StatsResponse, error) {

	resp := &api.StatsResponse{
		User: []api.UserStats{},
	}

	err := s.trm.Do(ctx, func(ctx context.Context) error {
		userStats, err := s.statsProvider.GetAssignmentsCountStats(ctx, sort)
		if err != nil {
			return err
		}
		prStats, err := s.statsProvider.GetPrStats(ctx)
		if err != nil {
			return err
		}

		for _, u := range userStats {
			stat := api.UserStats(*u)
			resp.User = append(resp.User, stat)
		}
		resp.Pr = api.PrStats(*prStats)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
