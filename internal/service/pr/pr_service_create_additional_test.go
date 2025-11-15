package pr_test

import (
	"context"
	"errors"
	"testing"

	"avito-intership-2025/internal/models"
	"avito-intership-2025/internal/service/mocks"
	"avito-intership-2025/internal/service/pr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullRequestService_Create_HappyPath_TwoReviewers(t *testing.T) {
	ctx := context.Background()

	// Inputs
	prID := "pr-happy-2"
	prName := "Feature A"
	authorID := "u1"
	teamID := 42

	// Mocks
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	author := &models.User{ID: authorID, TeamID: teamID}
	activeUsers := []string{"u2", "u3", "u4", authorID} // include author; must be excluded

	prCtrl.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).
		Return(prID, nil).
		Once()

	userGetter.On("GetById", ctx, authorID).Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, teamID).Return(activeUsers, nil).Once()

	// Two different calls for two reviewers; order is random, so we match by type
	reviewerProv.On("AssignReviewer", ctx, prID, mock.AnythingOfType("string")).Return(nil).Twice()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).Return(nil).Once()

	// SUT
	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Create(ctx, prID, prName, authorID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, prID, resp.ID)
	assert.Equal(t, prName, resp.Name)
	assert.Equal(t, authorID, resp.AuthorID)
	assert.Equal(t, pr.StatusOpen, resp.Status)
	assert.Len(t, resp.AssignedReviewers, 2)
	// Author must not be in reviewers and reviewers must be unique
	seen := map[string]struct{}{}
	for _, r := range resp.AssignedReviewers {
		assert.NotEqual(t, authorID, r)
		_, existed := seen[r]
		assert.False(t, existed, "reviewers should be unique")
		seen[r] = struct{}{}
	}
}

func TestPullRequestService_Create_Error_CreateFails(t *testing.T) {
	ctx := context.Background()

	prID := "ignored"
	prName := "Create fails"
	authorID := "u1"
	teamID := 77

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	author := &models.User{ID: authorID, TeamID: teamID}
	activeUsers := []string{"u2"}

	userGetter.On("GetById", ctx, authorID).Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, teamID).Return(activeUsers, nil).Once()

	createErr := errors.New("create error")
	prCtrl.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).
		Return("", createErr).
		Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, createErr, err)
		}).Return(createErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Create(ctx, prID, prName, authorID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, createErr, err)
	reviewerProv.AssertNotCalled(t, "AssignReviewer", mock.Anything, mock.Anything, mock.Anything)
}

func TestPullRequestService_Create_Error_GetAuthor(t *testing.T) {
	ctx := context.Background()

	prID := "pr-get-author"
	prName := "Author lookup fails"
	authorID := "u1"
	teamID := 100 // не используется, но пусть останется

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	// Create НЕ должен вызываться, поэтому не задаём ожидание на prCtrl.Create

	getErr := errors.New("author not found")
	userGetter.On("GetById", ctx, authorID).
		Return((*models.User)(nil), getErr).
		Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, getErr, err)
		}).Return(getErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Create(ctx, prID, prName, authorID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, getErr, err)

	// Проверяем, что этих вызовов не было
	prCtrl.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	reviewerProv.AssertNotCalled(t, "AssignReviewer", mock.Anything, mock.Anything, mock.Anything)
	userGetter.AssertNotCalled(t, "GetActiveUsersIDInTeam", mock.Anything, teamID)
}

func TestPullRequestService_Create_Error_AssignFirst(t *testing.T) {
	ctx := context.Background()

	prID := "pr-assign-first"
	prName := "Assign first fails"
	authorID := "u1"
	teamID := 7

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	author := &models.User{ID: authorID, TeamID: teamID}
	activeUsers := []string{"u2", "u3", authorID}

	prCtrl.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).
		Return(prID, nil).
		Once()

	userGetter.On("GetById", ctx, authorID).Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, teamID).Return(activeUsers, nil).Once()

	assignErr := errors.New("assign failed")
	reviewerProv.On("AssignReviewer", ctx, prID, mock.AnythingOfType("string")).
		Return(assignErr).
		Once() // fail on first assignment; second never called

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, assignErr, err)
		}).Return(assignErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Create(ctx, prID, prName, authorID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, assignErr, err)
	// Ensure there wasn't a second attempt registered
	reviewerProv.AssertNumberOfCalls(t, "AssignReviewer", 1)
}

func TestPullRequestService_Create_HappyPath_OneReviewer(t *testing.T) {
	ctx := context.Background()

	prID := "pr-one"
	prName := "Only one candidate"
	authorID := "u1"
	teamID := 2

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	author := &models.User{ID: authorID, TeamID: teamID}
	activeUsers := []string{"u2", authorID} // only one non-author candidate

	prCtrl.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).
		Return(prID, nil).
		Once()

	userGetter.On("GetById", ctx, authorID).Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, teamID).Return(activeUsers, nil).Once()

	reviewerProv.On("AssignReviewer", ctx, prID, mock.AnythingOfType("string")).Return(nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).Return(nil).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Create(ctx, prID, prName, authorID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, prID, resp.ID)
	assert.Equal(t, prName, resp.Name)
	assert.Equal(t, authorID, resp.AuthorID)
	assert.Equal(t, pr.StatusOpen, resp.Status)
	assert.Len(t, resp.AssignedReviewers, 1)
	assert.NotEqual(t, authorID, resp.AssignedReviewers[0])
}
