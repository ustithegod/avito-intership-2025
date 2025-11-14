package user_test

import (
	"avito-intership-2025/internal/models"
	"avito-intership-2025/internal/service/mocks"
	u "avito-intership-2025/internal/service/user"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_SetIsActive_Success(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockPrProvider := mocks.NewPrProvider(t) // not used here, but required by constructor
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

	// Expectations inside transaction
	mockUserChanger.On("SetIsActive", ctx, userID, isActive).Return(nil).Once()
	mockUserChanger.On("GetById", ctx, userID).Return(user, nil).Once()
	mockTeamIDProvider.On("GetTeamNameByID", ctx, 42).Return(teamName, nil).Once()

	// Transaction manager should execute provided function and return nil
	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).
		Return(nil).
		Once()

	service := u.NewUserService(mockTRM, mockPrProvider, mockUserChanger, mockTeamIDProvider)
	resp, err := service.SetIsActive(ctx, userID, isActive)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
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

	// First call in tx fails
	mockUserChanger.On("SetIsActive", ctx, userID, isActive).Return(dbErr).Once()

	// Transaction should propagate error from inner function
	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, dbErr))
		}).
		Return(dbErr).
		Once()

	service := u.NewUserService(mockTRM, nil, mockUserChanger, nil)
	resp, err := service.SetIsActive(ctx, userID, isActive)

	assert.Nil(t, resp)
	assert.Error(t, err)
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

	mockUserChanger.On("SetIsActive", ctx, userID, isActive).Return(nil).Once()
	mockUserChanger.On("GetById", ctx, userID).Return((*models.User)(nil), dbErr).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, dbErr))
		}).
		Return(dbErr).
		Once()

	service := u.NewUserService(mockTRM, nil, mockUserChanger, nil)
	resp, err := service.SetIsActive(ctx, userID, isActive)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, dbErr))
}

func TestUserService_SetIsActive_GetTeamNameByIDError(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockUserChanger := mocks.NewUserChanger(t)
	mockTeamIDProvider := mocks.NewTeamIDProvider(t)

	userID := "u123"
	isActive := true
	teamID := 100
	dbErr := errors.New("team not found")

	mockUserChanger.On("SetIsActive", ctx, userID, isActive).Return(nil).Once()
	mockUserChanger.On("GetById", ctx, userID).
		Return(&models.User{ID: userID, Name: "Bob", TeamID: teamID, IsActive: isActive}, nil).
		Once()
	mockTeamIDProvider.On("GetTeamNameByID", ctx, teamID).Return("", dbErr).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, dbErr))
		}).
		Return(dbErr).
		Once()

	service := u.NewUserService(mockTRM, nil, mockUserChanger, mockTeamIDProvider)
	resp, err := service.SetIsActive(ctx, userID, isActive)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, dbErr))
}

func TestUserService_GetReview_Success_WithPRs(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockPrProvider := mocks.NewPrProvider(t)
	mockUserChanger := mocks.NewUserChanger(t)

	userID := "u456"
	prs := []*models.PullRequest{
		{
			ID:       "pr-1",
			Title:    "Fix login",
			AuthorId: userID,
			Status:   "OPEN",
		},
		{
			ID:       "pr-2",
			Title:    "Add tests",
			AuthorId: userID,
			Status:   "MERGED",
		},
	}

	// Expectations inside transaction
	mockUserChanger.On("GetById", ctx, userID).Return(&models.User{ID: userID}, nil).Once()
	mockPrProvider.On("GetUserReviews", ctx, userID).Return(prs, nil).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).
		Return(nil).
		Once()

	service := u.NewUserService(mockTRM, mockPrProvider, mockUserChanger, nil)
	resp, err := service.GetReview(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userID, resp.UserID)
	assert.Len(t, resp.PullRequests, 2)

	assert.Equal(t, "pr-1", resp.PullRequests[0].ID)
	assert.Equal(t, "Fix login", resp.PullRequests[0].Name)
	assert.Equal(t, userID, resp.PullRequests[0].AuthorID)
	assert.Equal(t, "OPEN", resp.PullRequests[0].Status)

	assert.Equal(t, "pr-2", resp.PullRequests[1].ID)
	assert.Equal(t, "Add tests", resp.PullRequests[1].Name)
	assert.Equal(t, "MERGED", resp.PullRequests[1].Status)
}

func TestUserService_GetReview_Success_NoPRs(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockPrProvider := mocks.NewPrProvider(t)
	mockUserChanger := mocks.NewUserChanger(t)

	userID := "u789"

	mockUserChanger.On("GetById", ctx, userID).Return(&models.User{ID: userID}, nil).Once()
	mockPrProvider.On("GetUserReviews", ctx, userID).Return([]*models.PullRequest{}, nil).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).
		Return(nil).
		Once()

	service := u.NewUserService(mockTRM, mockPrProvider, mockUserChanger, nil)
	resp, err := service.GetReview(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, userID, resp.UserID)
	assert.Empty(t, resp.PullRequests)
}

func TestUserService_GetReview_GetByIdError(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockPrProvider := mocks.NewPrProvider(t) // not used due to early error
	mockUserChanger := mocks.NewUserChanger(t)

	userID := "u999"
	dbErr := errors.New("user not found")

	mockUserChanger.On("GetById", ctx, userID).Return((*models.User)(nil), dbErr).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, dbErr))
		}).
		Return(dbErr).
		Once()

	service := u.NewUserService(mockTRM, mockPrProvider, mockUserChanger, nil)
	resp, err := service.GetReview(ctx, userID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, dbErr))
}

func TestUserService_GetReview_PrProviderError(t *testing.T) {
	ctx := context.Background()

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	mockPrProvider := mocks.NewPrProvider(t)
	mockUserChanger := mocks.NewUserChanger(t)

	userID := "u111"
	prErr := errors.New("reviews unreachable")

	mockUserChanger.On("GetById", ctx, userID).Return(&models.User{ID: userID}, nil).Once()
	mockPrProvider.On("GetUserReviews", ctx, userID).Return(([]*models.PullRequest)(nil), prErr).Once()

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, prErr))
		}).
		Return(prErr).
		Once()

	service := u.NewUserService(mockTRM, mockPrProvider, mockUserChanger, nil)
	resp, err := service.GetReview(ctx, userID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, prErr))
}
