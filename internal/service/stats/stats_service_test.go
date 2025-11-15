package stats_test

import (
	"context"
	"errors"
	"testing"

	"avito-intership-2025/internal/models"
	"avito-intership-2025/internal/service/mocks"
	"avito-intership-2025/internal/service/stats"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatsService_GetStatistics_Success(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockStatsProvider := mocks.NewStatsProvider(t)

	sort := "desc"

	userStats := []*models.UserStatistics{
		{
			UserID:          "user1",
			Username:        "Alice",
			AssignmentCount: 15,
		},
		{
			UserID:          "user2",
			Username:        "Bob",
			AssignmentCount: 10,
		},
		{
			UserID:          "user3",
			Username:        "Charlie",
			AssignmentCount: 5,
		},
	}

	prStats := &models.PrStatistics{
		PrCount:   100,
		OpenPrs:   30,
		MergedPrs: 70,
	}

	// Expectations inside transaction
	mockStatsProvider.On("GetAssignmentsCountStats", ctx, sort).Return(userStats, nil).Once()
	mockStatsProvider.On("GetPrStats", ctx).Return(prStats, nil).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).
		Return(nil).
		Once()

	service := stats.NewStatsService(mockTRM, mockStatsProvider)
	resp, err := service.GetStatistics(ctx, sort)

	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify user stats
	assert.Len(t, resp.User, 3)

	assert.Equal(t, "user1", resp.User[0].UserID)
	assert.Equal(t, "Alice", resp.User[0].Username)
	assert.Equal(t, 15, resp.User[0].AssignmentCount)

	assert.Equal(t, "user2", resp.User[1].UserID)
	assert.Equal(t, "Bob", resp.User[1].Username)
	assert.Equal(t, 10, resp.User[1].AssignmentCount)

	assert.Equal(t, "user3", resp.User[2].UserID)
	assert.Equal(t, "Charlie", resp.User[2].Username)
	assert.Equal(t, 5, resp.User[2].AssignmentCount)

	// Verify PR stats
	assert.Equal(t, 100, resp.Pr.PrCount)
	assert.Equal(t, 30, resp.Pr.OpenPrs)
	assert.Equal(t, 70, resp.Pr.MergedPrs)
}

func TestStatsService_GetStatistics_EmptyUserStats(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockStatsProvider := mocks.NewStatsProvider(t)

	sort := "asc"

	userStats := []*models.UserStatistics{}
	prStats := &models.PrStatistics{
		PrCount:   0,
		OpenPrs:   0,
		MergedPrs: 0,
	}

	mockStatsProvider.On("GetAssignmentsCountStats", ctx, sort).Return(userStats, nil).Once()
	mockStatsProvider.On("GetPrStats", ctx).Return(prStats, nil).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).
		Return(nil).
		Once()

	service := stats.NewStatsService(mockTRM, mockStatsProvider)
	resp, err := service.GetStatistics(ctx, sort)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.User)
	assert.Equal(t, 0, resp.Pr.PrCount)
	assert.Equal(t, 0, resp.Pr.OpenPrs)
	assert.Equal(t, 0, resp.Pr.MergedPrs)
}

func TestStatsService_GetStatistics_GetAssignmentsCountStatsError(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockStatsProvider := mocks.NewStatsProvider(t)

	sort := "desc"
	dbErr := errors.New("database connection failed")

	mockStatsProvider.On("GetAssignmentsCountStats", ctx, sort).Return(([]*models.UserStatistics)(nil), dbErr).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.ErrorIs(t, err, dbErr)
		}).
		Return(dbErr).
		Once()

	service := stats.NewStatsService(mockTRM, mockStatsProvider)
	resp, err := service.GetStatistics(ctx, sort)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.ErrorIs(t, err, dbErr)
}

func TestStatsService_GetStatistics_GetPrStatsError(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockStatsProvider := mocks.NewStatsProvider(t)

	sort := "asc"
	dbErr := errors.New("failed to fetch PR stats")

	userStats := []*models.UserStatistics{
		{
			UserID:          "user1",
			Username:        "Dave",
			AssignmentCount: 3,
		},
	}

	mockStatsProvider.On("GetAssignmentsCountStats", ctx, sort).Return(userStats, nil).Once()
	mockStatsProvider.On("GetPrStats", ctx).Return((*models.PrStatistics)(nil), dbErr).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.ErrorIs(t, err, dbErr)
		}).
		Return(dbErr).
		Once()

	service := stats.NewStatsService(mockTRM, mockStatsProvider)
	resp, err := service.GetStatistics(ctx, sort)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.ErrorIs(t, err, dbErr)
}
