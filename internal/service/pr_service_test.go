package service

import (
	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/models"
	repo "avito-intership-2025/internal/repository"
	"avito-intership-2025/internal/service/mocks"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullRequestService_Create(t *testing.T) {
	ctx := context.Background()
	prID := "pr-123"
	prName := "Test PR"
	authorID := "user-123"
	teamID := 1

	tests := []struct {
		name       string
		setupMocks func(
			*mocks.PrController,
			*mocks.UserGetter,
			*mocks.ReviewerProvider,
			*mocks.MockManager,
		)
		checkResponse func(*testing.T, *api.PullRequestSchema, error)
	}{
		{
			name: "successful creation with reviewers",
			setupMocks: func(
				prCtrl *mocks.PrController,
				userGetter *mocks.UserGetter,
				reviewerProv *mocks.ReviewerProvider,
				trm *mocks.MockManager,
			) {
				author := &models.User{ID: authorID, TeamID: teamID}
				activeUsers := []string{"user-456", "user-789", "user-999"}

				prCtrl.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).
					Return(prID, nil)
				userGetter.On("GetById", ctx, authorID).Return(author, nil)
				userGetter.On("GetActiveUsersIDInTeam", ctx, teamID).Return(activeUsers, nil)
				// Используем mock.Anything для ревьюверов, так как порядок случайный
				reviewerProv.On("AssignReviewer", ctx, prID, mock.AnythingOfType("string")).Return(nil).Twice()
				trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					}).Return(nil)
			},
			checkResponse: func(t *testing.T, resp *api.PullRequestSchema, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, prID, resp.ID)
				assert.Equal(t, prName, resp.Name)
				assert.Equal(t, authorID, resp.AuthorID)
				assert.Equal(t, StatusOpen, resp.Status)
				assert.Len(t, resp.AssignedReviewers, 2)
				// Проверяем, что автор не среди ревьюверов
				for _, reviewer := range resp.AssignedReviewers {
					assert.NotEqual(t, authorID, reviewer)
				}
				// Проверяем, что все ревьюверы уникальны
				assert.Equal(t, len(resp.AssignedReviewers), len(unique(resp.AssignedReviewers)))
			},
		},
		{
			name: "creation fails when pr creation fails",
			setupMocks: func(
				prCtrl *mocks.PrController,
				userGetter *mocks.UserGetter,
				reviewerProv *mocks.ReviewerProvider,
				trm *mocks.MockManager,
			) {
				prCtrl.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).
					Return("", errors.New("create error"))
				trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					}).Return(errors.New("create error"))
			},
			checkResponse: func(t *testing.T, resp *api.PullRequestSchema, err error) {
				assert.Error(t, err)
				assert.Equal(t, "create error", err.Error())
				assert.Nil(t, resp)
			},
		},
		{
			name: "creation fails when author not found",
			setupMocks: func(
				prCtrl *mocks.PrController,
				userGetter *mocks.UserGetter,
				reviewerProv *mocks.ReviewerProvider,
				trm *mocks.MockManager,
			) {
				prCtrl.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).
					Return(prID, nil)
				userGetter.On("GetById", ctx, authorID).Return(nil, errors.New("user not found"))
				trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					}).Return(errors.New("user not found"))
			},
			checkResponse: func(t *testing.T, resp *api.PullRequestSchema, err error) {
				assert.Error(t, err)
				assert.Equal(t, "user not found", err.Error())
				assert.Nil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			prCtrl := mocks.NewPrController(t)
			userGetter := mocks.NewUserGetter(t)
			reviewerProv := mocks.NewReviewerProvider(t)
			trm := &mocks.MockManager{}
			trm.Test(t)
			t.Cleanup(func() { trm.AssertExpectations(t) })

			tt.setupMocks(prCtrl, userGetter, reviewerProv, trm)

			// Create service
			service := &PullRequestService{
				prController:     prCtrl,
				userGetter:       userGetter,
				reviewerProvider: reviewerProv,
				trm:              trm,
			}

			// Execute
			resp, err := service.Create(ctx, prID, prName, authorID)

			// Check response
			tt.checkResponse(t, resp, err)

			// Verify all expectations
			prCtrl.AssertExpectations(t)
			userGetter.AssertExpectations(t)
			reviewerProv.AssertExpectations(t)
			trm.AssertExpectations(t)
		})
	}
}

func TestPullRequestService_Reassign(t *testing.T) {
	ctx := context.Background()
	prID := "pr-123"
	oldRev := "user-456"

	tests := []struct {
		name       string
		setupMocks func(
			*mocks.PrController,
			*mocks.UserGetter,
			*mocks.ReviewerProvider,
			*mocks.MockManager,
		)
		checkResponse func(*testing.T, *api.ReassignResponse, error)
	}{
		{
			name: "successful reassign",
			setupMocks: func(
				prCtrl *mocks.PrController,
				userGetter *mocks.UserGetter,
				reviewerProv *mocks.ReviewerProvider,
				trm *mocks.MockManager,
			) {
				pr := &models.PullRequest{
					ID:       prID,
					Title:    "Test PR",
					AuthorId: "user-123",
					Status:   StatusOpen,
				}
				author := &models.User{ID: "user-123", TeamID: 1}
				activeUsers := []string{"user-456", "user-789", "user-999", "user-111"}
				assignedReviewers := []string{"user-456", "user-789"}
				finalReviewers := []string{"user-999", "user-789"} // новый ревьювер заменил старого

				prCtrl.On("GetById", ctx, prID).Return(pr, nil).Twice()
				userGetter.On("GetById", ctx, "user-123").Return(author, nil)
				userGetter.On("GetActiveUsersIDInTeam", ctx, 1).Return(activeUsers, nil)
				reviewerProv.On("GetPrReviewers", ctx, prID).Return(assignedReviewers, nil).Once()
				// Используем mock.Anything для нового ревьювера
				reviewerProv.On("ReassignReviewer", ctx, prID, oldRev, mock.AnythingOfType("string")).
					Return(nil)
				reviewerProv.On("GetPrReviewers", ctx, prID).Return(finalReviewers, nil).Once()
				trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					}).Return(nil)
			},
			checkResponse: func(t *testing.T, resp *api.ReassignResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, prID, resp.PullRequest.ID)
				assert.Equal(t, "Test PR", resp.PullRequest.Name)
				assert.Equal(t, "user-123", resp.PullRequest.AuthorID)
				assert.Equal(t, StatusOpen, resp.PullRequest.Status)
				assert.Len(t, resp.PullRequest.AssignedReviewers, 2)
				assert.NotEmpty(t, resp.ReplacedBy)
				assert.NotEqual(t, oldRev, resp.ReplacedBy)
				// Проверяем, что старый ревьювер больше не в списке
				assert.NotContains(t, resp.PullRequest.AssignedReviewers, oldRev)
			},
		},
		{
			name: "reassign fails when PR is merged",
			setupMocks: func(
				prCtrl *mocks.PrController,
				userGetter *mocks.UserGetter,
				reviewerProv *mocks.ReviewerProvider,
				trm *mocks.MockManager,
			) {
				pr := &models.PullRequest{
					ID:       prID,
					Title:    "Test PR",
					AuthorId: "user-123",
					Status:   StatusMerged,
				}

				prCtrl.On("GetById", ctx, prID).Return(pr, nil)
				trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						err := fn(ctx)
						assert.Error(t, err)
						assert.Equal(t, ErrTryMergeMerged, err)
					}).Return(ErrTryMergeMerged)
			},
			checkResponse: func(t *testing.T, resp *api.ReassignResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, ErrTryMergeMerged, err)
				assert.Nil(t, resp)
			},
		},
		{
			name: "reassign fails when reviewer not assigned",
			setupMocks: func(
				prCtrl *mocks.PrController,
				userGetter *mocks.UserGetter,
				reviewerProv *mocks.ReviewerProvider,
				trm *mocks.MockManager,
			) {
				pr := &models.PullRequest{
					ID:       prID,
					Title:    "Test PR",
					AuthorId: "user-123",
					Status:   StatusOpen,
				}
				assignedReviewers := []string{"user-789"} // oldRev не в списке

				prCtrl.On("GetById", ctx, prID).Return(pr, nil).Once()
				reviewerProv.On("GetPrReviewers", ctx, prID).Return(assignedReviewers, nil).Once()
				trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					}).Return(repo.ErrNotAssigned)
			},
			checkResponse: func(t *testing.T, resp *api.ReassignResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, repo.ErrNotAssigned, err)
				assert.Nil(t, resp)
			},
		},
		{
			name: "reassign fails when no available users",
			setupMocks: func(
				prCtrl *mocks.PrController,
				userGetter *mocks.UserGetter,
				reviewerProv *mocks.ReviewerProvider,
				trm *mocks.MockManager,
			) {
				pr := &models.PullRequest{
					ID:       prID,
					Title:    "Test PR",
					AuthorId: "user-123",
					Status:   StatusOpen,
				}
				author := &models.User{ID: "user-123", TeamID: 1}
				activeUsers := []string{"user-123", "user-456"} // только автор и старый ревьювер
				assignedReviewers := []string{"user-456"}

				prCtrl.On("GetById", ctx, prID).Return(pr, nil).Once()
				userGetter.On("GetById", ctx, "user-123").Return(author, nil)
				userGetter.On("GetActiveUsersIDInTeam", ctx, 1).Return(activeUsers, nil)
				reviewerProv.On("GetPrReviewers", ctx, prID).Return(assignedReviewers, nil).Once()
				trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						fn(ctx)
					}).Return(repo.ErrNoCandidate)
			},
			checkResponse: func(t *testing.T, resp *api.ReassignResponse, err error) {
				assert.Error(t, err)
				assert.Equal(t, repo.ErrNoCandidate, err)
				assert.Nil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			prCtrl := mocks.NewPrController(t)
			userGetter := mocks.NewUserGetter(t)
			reviewerProv := mocks.NewReviewerProvider(t)
			trm := &mocks.MockManager{}
			trm.Test(t)
			t.Cleanup(func() { trm.AssertExpectations(t) })

			tt.setupMocks(prCtrl, userGetter, reviewerProv, trm)

			// Create service
			service := &PullRequestService{
				prController:     prCtrl,
				userGetter:       userGetter,
				reviewerProvider: reviewerProv,
				trm:              trm,
			}

			// Execute
			resp, err := service.Reassign(ctx, prID, oldRev)

			// Check response
			tt.checkResponse(t, resp, err)

			// Verify all expectations
			prCtrl.AssertExpectations(t)
			userGetter.AssertExpectations(t)
			reviewerProv.AssertExpectations(t)
			trm.AssertExpectations(t)
		})
	}
}

// Вспомогательная функция для проверки уникальности
func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
