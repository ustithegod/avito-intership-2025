package team_test

import (
	"context"
	"errors"
	"testing"

	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/models"
	repo "avito-intership-2025/internal/repository"
	"avito-intership-2025/internal/service/mocks"
	"avito-intership-2025/internal/service/team"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTeamService_Add_Success(t *testing.T) {
	ctx := context.Background()

	mockTeamProvider := mocks.NewTeamProvider(t)

	mockUserProvider := mocks.NewUserProvider(t)

	mockTRM := &mocks.MockManager{}

	mockTRM.Test(t)

	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	teamName := "avengers"

	users := []api.TeamMember{

		{UserID: "u1", Username: "Tony", IsActive: true},

		{UserID: "u2", Username: "Steve", IsActive: false},
	}

	teamID := 42

	mockTeamProvider.On("Create", ctx, teamName).Return(teamID, nil)

	mockUserProvider.On("Save", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == "u1" && u.Name == "Tony" && u.TeamID == teamID && u.IsActive
	})).Return("", nil)

	mockUserProvider.On("Save", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == "u2" && u.Name == "Steve" && u.TeamID == teamID && !u.IsActive
	})).Return("", nil)

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)

			assert.NoError(t, fn(ctx))
		}).
		Return(nil).Once()

	service := team.NewTeamService(mockTRM, mockTeamProvider, mockUserProvider)

	resp, err := service.Add(ctx, teamName, users)

	assert.NoError(t, err)

	assert.Equal(t, teamName, resp.TeamName)

	assert.Len(t, resp.Members, 2)

	assert.Equal(t, "Tony", resp.Members[0].Username)

	assert.True(t, resp.Members[0].IsActive)
}

func TestTeamService_Add_TeamExists(t *testing.T) {
	ctx := context.Background()

	mockTeamProvider := mocks.NewTeamProvider(t)

	mockTRM := &mocks.MockManager{}

	mockTRM.Test(t)

	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	teamName := "avengers"

	users := []api.TeamMember{{UserID: "u1", Username: "Thor", IsActive: true}}

	// Create ВЫЗЫВАЕТСЯ и возвращает ErrTeamExists

	mockTeamProvider.On("Create", ctx, teamName).Return(0, repo.ErrTeamExists)

	// trm.Do выполняет функцию и получает ошибку

	mockTRM.On("Do", ctx, mock.Anything).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)

			err := fn(ctx)

			assert.ErrorIs(t, err, repo.ErrTeamExists)
		}).
		Return(repo.ErrTeamExists).
		Once()

	service := team.NewTeamService(mockTRM, mockTeamProvider, nil)

	resp, err := service.Add(ctx, teamName, users)

	assert.Nil(t, resp)

	assert.ErrorIs(t, err, repo.ErrTeamExists)
}

func TestTeamService_Get_Success(t *testing.T) {
	ctx := context.Background()

	mockTeamProvider := mocks.NewTeamProvider(t)

	mockUserProvider := mocks.NewUserProvider(t)

	teamName := "x-men"

	tm := &models.Team{ID: 10, Name: teamName}

	users := []*models.User{

		{ID: "wolverine", Name: "Logan", TeamID: 10, IsActive: true},

		{ID: "storm", Name: "Ororo", TeamID: 10, IsActive: false},
	}

	mockTeamProvider.On("GetByTeamName", ctx, teamName).Return(tm, nil)

	mockUserProvider.On("GetUsersInTeam", ctx, teamName).Return(users, nil)

	service := team.NewTeamService(nil, mockTeamProvider, mockUserProvider)

	resp, err := service.Get(ctx, teamName)

	assert.NoError(t, err)

	assert.Equal(t, teamName, resp.TeamName)

	assert.Len(t, resp.Members, 2)
}

func TestTeamService_Get_NotFound(t *testing.T) {
	ctx := context.Background()

	mockTeamProvider := mocks.NewTeamProvider(t)

	teamName := "fantastic-four"

	mockTeamProvider.On("GetByTeamName", ctx, teamName).Return((*models.Team)(nil), repo.ErrNotFound)

	service := team.NewTeamService(nil, mockTeamProvider, nil)

	resp, err := service.Get(ctx, teamName)

	assert.Nil(t, resp)

	assert.ErrorIs(t, err, repo.ErrNotFound)
}

func TestTeamService_Add_SaveError(t *testing.T) {
	ctx := context.Background()

	mockTeamProvider := mocks.NewTeamProvider(t)
	mockUserProvider := mocks.NewUserProvider(t)

	mockTRM := &mocks.MockManager{}
	mockTRM.Test(t)
	t.Cleanup(func() { mockTRM.AssertExpectations(t) })

	teamName := "justice-league"
	users := []api.TeamMember{
		{UserID: "u1", Username: "Bruce", IsActive: true},
		{UserID: "u2", Username: "Clark", IsActive: true},
	}
	teamID := 7
	saveErr := errors.New("save failed")

	mockTeamProvider.On("Create", ctx, teamName).Return(teamID, nil)

	mockUserProvider.On("Save", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == "u1" && u.Name == "Bruce" && u.TeamID == teamID && u.IsActive
	})).Return("", nil)

	mockUserProvider.On("Save", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == "u2" && u.Name == "Clark" && u.TeamID == teamID && u.IsActive
	})).Return("", saveErr)

	mockTRM.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.ErrorIs(t, err, saveErr)
		}).
		Return(saveErr).
		Once()

	service := team.NewTeamService(mockTRM, mockTeamProvider, mockUserProvider)
	resp, err := service.Add(ctx, teamName, users)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.ErrorIs(t, err, saveErr)
}

func TestTeamService_Get_GetUsersError(t *testing.T) {
	ctx := context.Background()

	mockTeamProvider := mocks.NewTeamProvider(t)
	mockUserProvider := mocks.NewUserProvider(t)

	teamName := "x-men-err"
	tm := &models.Team{ID: 10, Name: teamName}
	getErr := errors.New("users fetch failed")

	mockTeamProvider.On("GetByTeamName", ctx, teamName).Return(tm, nil)
	mockUserProvider.On("GetUsersInTeam", ctx, teamName).Return(([]*models.User)(nil), getErr)

	service := team.NewTeamService(nil, mockTeamProvider, mockUserProvider)
	resp, err := service.Get(ctx, teamName)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.ErrorIs(t, err, getErr)
}
