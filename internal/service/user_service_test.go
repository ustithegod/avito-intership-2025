package service

import (
	"avito-intership-2025/internal/models"
	"avito-intership-2025/internal/service/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_SetIsActive_Success(t *testing.T) {
	ctx := context.Background()

	// Моки
	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockPrProvider := mocks.NewPrProvider(t)
	mockUserChanger := mocks.NewUserChanger(t)
	mockTeamIDProvider := mocks.NewTeamIDProvider(t)

	userID := "u123"
	isActive := true
	user := &models.User{
		ID:       userID,
		Name:     "Alice",
		TeamID:   42,
		IsActive: isActive,
	}
	teamName := "backend"

	// Ожидания
	mockUserChanger.On("SetIsActive", ctx, userID, isActive).Return(nil)
	mockUserChanger.On("GetById", ctx, userID).Return(user, nil)
	mockTeamIDProvider.On("GetTeamNameByID", ctx, 42).Return(teamName, nil)

	mockTRM.On("Do", ctx, mock.Anything).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).
		Return(nil).
		Once()

	// Act
	service := NewUserService(mockTRM, mockPrProvider, mockUserChanger, mockTeamIDProvider)
	resp, err := service.SetIsActive(ctx, userID, isActive)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, "Alice", resp.Username)
	assert.Equal(t, teamName, resp.TeamName)
	assert.True(t, resp.IsActive)
}

func TestUserService_SetIsActive_SetIsActiveError(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockUserChanger := mocks.NewUserChanger(t)

	userID := "u123"
	isActive := false
	dbErr := errors.New("failed to update")

	mockUserChanger.On("SetIsActive", ctx, userID, isActive).Return(dbErr)

	mockTRM.On("Do", ctx, mock.Anything).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.True(t, errors.Is(err, dbErr))
		}).
		Return(dbErr).
		Once()

	service := NewUserService(mockTRM, nil, mockUserChanger, nil)
	resp, err := service.SetIsActive(ctx, userID, isActive)

	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, dbErr))
}

func TestUserService_SetIsActive_GetByIdError(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockUserChanger := mocks.NewUserChanger(t)

	userID := "u123"
	isActive := true
	dbErr := errors.New("user not found")

	mockUserChanger.On("SetIsActive", ctx, userID, isActive).Return(nil)
	mockUserChanger.On("GetById", ctx, userID).Return((*models.User)(nil), dbErr)

	mockTRM.On("Do", ctx, mock.Anything).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.True(t, errors.Is(err, dbErr))
		}).
		Return(dbErr).
		Once()

	service := NewUserService(mockTRM, nil, mockUserChanger, nil)
	resp, err := service.SetIsActive(ctx, userID, isActive)

	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, dbErr))
}

func TestUserService_GetReview_Success(t *testing.T) {
	ctx := context.Background()

	mockPrProvider := mocks.NewPrProvider(t)

	userID := "u456"
	now := time.Now()
	prs := []*models.PullRequest{
		{
			ID:                "pr-1",
			Title:             "Fix login",
			AuthorId:          userID,
			Status:            "OPEN", // ← ВЕРХНИЙ РЕГИСТР
			NeedMoreReviewers: false,
			CreatedAt:         &now,
		},
		{
			ID:                "pr-2",
			Title:             "Add tests",
			AuthorId:          userID,
			Status:            "MERGED", // ← ВЕРХНИЙ РЕГИСТР
			NeedMoreReviewers: true,
			CreatedAt:         &now,
		},
	}

	mockPrProvider.On("GetUserReviews", ctx, userID).Return(prs, nil)

	service := NewUserService(nil, mockPrProvider, nil, nil)
	resp, err := service.GetReview(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, userID, resp.UserID)
	assert.Len(t, resp.PullRequests, 2)

	// Проверяем порядок
	assert.Equal(t, "pr-1", resp.PullRequests[0].ID)
	assert.Equal(t, "Fix login", resp.PullRequests[0].Name)
	assert.Equal(t, userID, resp.PullRequests[0].AuthorID)
	assert.Equal(t, "OPEN", resp.PullRequests[0].Status) // ← проверяем регистр

	assert.Equal(t, "pr-2", resp.PullRequests[1].ID)
	assert.Equal(t, "Add tests", resp.PullRequests[1].Name)
	assert.Equal(t, "MERGED", resp.PullRequests[1].Status)
}

func TestUserService_GetReview_NoPRs(t *testing.T) {
	ctx := context.Background()

	mockPrProvider := mocks.NewPrProvider(t)
	userID := "u789"

	mockPrProvider.On("GetUserReviews", ctx, userID).Return([]*models.PullRequest{}, nil)

	service := NewUserService(nil, mockPrProvider, nil, nil)
	resp, err := service.GetReview(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, userID, resp.UserID)
	assert.Empty(t, resp.PullRequests)
}

func TestUserService_GetReview_Error(t *testing.T) {
	ctx := context.Background()

	mockPrProvider := mocks.NewPrProvider(t)
	userID := "u999"
	prErr := errors.New("github unreachable")

	mockPrProvider.On("GetUserReviews", ctx, userID).Return(([]*models.PullRequest)(nil), prErr)

	service := NewUserService(nil, mockPrProvider, nil, nil)
	resp, err := service.GetReview(ctx, userID)

	assert.Nil(t, resp)
	assert.True(t, errors.Is(err, prErr))
}
