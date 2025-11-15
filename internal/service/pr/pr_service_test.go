package pr_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"avito-intership-2025/internal/models"
	repo "avito-intership-2025/internal/repository"
	"avito-intership-2025/internal/service/mocks"
	"avito-intership-2025/internal/service/pr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullRequestService_Create_Success_NoAvailableCandidates(t *testing.T) {
	ctx := context.Background()
	prID := "pr-1"
	prName := "No reviewers possible"
	authorID := "user-1"
	teamID := 10

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	author := &models.User{ID: authorID, TeamID: teamID}
	activeUsers := []string{authorID} // only author is active -> no candidates

	prCtrl.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).Return(prID, nil).Once()
	userGetter.On("GetById", ctx, authorID).Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, teamID).Return(activeUsers, nil).Once()
	// AssignReviewer should NOT be called because there are no candidates

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
	assert.Empty(t, resp.AssignedReviewers)
}

func TestPullRequestService_Create_AssignReviewerFailsOnSecond(t *testing.T) {
	ctx := context.Background()
	prID := "pr-2"
	prName := "Assign fail"
	authorID := "user-1"
	teamID := 11

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	author := &models.User{ID: authorID, TeamID: teamID}
	activeUsers := []string{"user-2", "user-3", "user-4"} // at least 2 candidates

	prCtrl.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).Return(prID, nil).Once()
	userGetter.On("GetById", ctx, authorID).Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, teamID).Return(activeUsers, nil).Once()

	assignErr := errors.New("assign failed")
	reviewerProv.
		On("AssignReviewer", ctx, prID, mock.AnythingOfType("string")).
		Return(nil).Once() // first assignment ok
	reviewerProv.
		On("AssignReviewer", ctx, prID, mock.AnythingOfType("string")).
		Return(assignErr).Once() // second assignment fails

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
}

func TestPullRequestService_Create_GetActiveUsersError(t *testing.T) {
	ctx := context.Background()
	prID := "pr-3"
	prName := "Active users error"
	authorID := "user-1"
	teamID := 12

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	author := &models.User{ID: authorID, TeamID: teamID}
	activeErr := errors.New("get active users failed")

	// Новый порядок: сначала GetById, потом GetActiveUsers... (падает),
	// Create НЕ должен вызываться
	userGetter.On("GetById", ctx, authorID).Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, teamID).Return(([]string)(nil), activeErr).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, activeErr, err)
		}).Return(activeErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Create(ctx, prID, prName, authorID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, activeErr, err)

	// Убедимся, что не вызывался Create и AssignReviewer
	prCtrl.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	reviewerProv.AssertNotCalled(t, "AssignReviewer", mock.Anything, mock.Anything, mock.Anything)
}

func TestPullRequestService_Merge_Success_FromOpen(t *testing.T) {
	ctx := context.Background()
	prID := "merge-1"

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	now := time.Now()
	open := &models.PullRequest{ID: prID, Title: "Merge me", AuthorId: "a1", Status: pr.StatusOpen}
	merged := &models.PullRequest{ID: prID, Title: "Merge me", AuthorId: "a1", Status: pr.StatusMerged, MergedAt: &now}
	reviewers := []string{"r1", "r2"}

	prCtrl.On("GetById", ctx, prID).Return(open, nil).Once()
	prCtrl.On("MarkAsMerged", ctx, prID).Return(nil).Once()
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(reviewers, nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).Return(nil).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, prID, resp.ID)
	assert.Equal(t, "Merge me", resp.Name)
	assert.Equal(t, pr.StatusMerged, resp.Status)
	assert.Equal(t, reviewers, resp.AssignedReviewers)
	assert.NotNil(t, resp.MergedAt)
}

func TestPullRequestService_Merge_Success_AlreadyMerged_NoMarkCall(t *testing.T) {
	ctx := context.Background()
	prID := "merge-2"

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	now := time.Now()
	merged := &models.PullRequest{
		ID:       prID,
		Title:    "Already merged",
		AuthorId: "a2",
		Status:   pr.StatusMerged,
		MergedAt: &now,
	}
	reviewers := []string{"r9"}

	// First GetById shows merged, so MarkAsMerged should not be needed
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	// The service requests it again after the (skipped) merge branch
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(reviewers, nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).Return(nil).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, pr.StatusMerged, resp.Status)
	assert.Equal(t, reviewers, resp.AssignedReviewers)
	// Ensure MarkAsMerged wasn't called
	prCtrl.AssertNotCalled(t, "MarkAsMerged", ctx, prID)
}

func TestPullRequestService_Merge_Error_FirstGetById(t *testing.T) {
	ctx := context.Background()
	prID := "merge-err-1"

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	getErr := errors.New("first get error")
	prCtrl.On("GetById", ctx, prID).Return((*models.PullRequest)(nil), getErr).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, getErr, err)
		}).Return(getErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, getErr, err)
}

func TestPullRequestService_Merge_Error_SecondGetById(t *testing.T) {
	ctx := context.Background()
	prID := "merge-err-2"

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	open := &models.PullRequest{ID: prID, Title: "X", AuthorId: "a1", Status: pr.StatusOpen}
	secondErr := errors.New("second get error")

	prCtrl.On("GetById", ctx, prID).Return(open, nil).Once()
	prCtrl.On("MarkAsMerged", ctx, prID).Return(nil).Once()
	prCtrl.On("GetById", ctx, prID).Return((*models.PullRequest)(nil), secondErr).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, secondErr, err)
		}).Return(secondErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, secondErr, err)
}

func TestPullRequestService_Merge_Error_GetReviewers(t *testing.T) {
	ctx := context.Background()
	prID := "merge-err-3"

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	merged := &models.PullRequest{ID: prID, Title: "X", AuthorId: "a1", Status: pr.StatusMerged}
	revErr := errors.New("reviewers error")

	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(([]string)(nil), revErr).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, revErr, err)
		}).Return(revErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, revErr, err)
}

func TestPullRequestService_Merge_Ignores_MarkAsMergedError(t *testing.T) {
	ctx := context.Background()
	prID := "merge-ign-1"

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	open := &models.PullRequest{ID: prID, Title: "X", AuthorId: "a1", Status: pr.StatusOpen}
	merged := &models.PullRequest{ID: prID, Title: "X", AuthorId: "a1", Status: pr.StatusMerged}
	reviewers := []string{"r1"}
	mergeErr := errors.New("merge failed") // Will be ignored by implementation

	prCtrl.On("GetById", ctx, prID).Return(open, nil).Once()
	prCtrl.On("MarkAsMerged", ctx, prID).Return(mergeErr).Once()
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(reviewers, nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			// Despite mark error, flow continues and returns success
			assert.NoError(t, fn(ctx))
		}).Return(nil).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, pr.StatusMerged, resp.Status)
	assert.Equal(t, reviewers, resp.AssignedReviewers)
}

func TestPullRequestService_Reassign_Success(t *testing.T) {
	ctx := context.Background()
	prID := "re-1"
	oldRev := "r1"

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	current := &models.PullRequest{ID: prID, Title: "Reassign", AuthorId: "a1", Status: pr.StatusOpen}
	author := &models.User{ID: "a1", TeamID: 77}
	active := []string{"a1", "r1", "r2", "r3"}
	assigned := []string{"r1", "r2"}
	final := []string{"r2", "r3"} // after reassign

	prCtrl.On("GetById", ctx, prID).Return(current, nil).Twice()
	userGetter.On("GetById", ctx, "a1").Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, 77).Return(active, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(assigned, nil).Once()
	reviewerProv.On("ReassignReviewer", ctx, prID, oldRev, mock.AnythingOfType("string")).Return(nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(final, nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			assert.NoError(t, fn(ctx))
		}).Return(nil).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Reassign(ctx, prID, oldRev)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, prID, resp.PullRequest.ID)
	assert.Equal(t, "Reassign", resp.PullRequest.Name)
	assert.Equal(t, pr.StatusOpen, resp.PullRequest.Status)
	assert.Len(t, resp.PullRequest.AssignedReviewers, 2)
	assert.NotEmpty(t, resp.ReplacedBy)
	assert.NotEqual(t, oldRev, resp.ReplacedBy)
	assert.NotContains(t, resp.PullRequest.AssignedReviewers, oldRev)
}

func TestPullRequestService_Reassign_NoCandidate(t *testing.T) {
	ctx := context.Background()
	prID := "re-2"
	oldRev := "b1"

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	current := &models.PullRequest{ID: prID, Title: "No cand", AuthorId: "a1", Status: pr.StatusOpen}
	author := &models.User{ID: "a1", TeamID: 5}
	active := []string{"a1", "b1"} // excluded = author + assigned = active -> no candidate
	assigned := []string{"b1"}

	prCtrl.On("GetById", ctx, prID).Return(current, nil).Once()
	userGetter.On("GetById", ctx, "a1").Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, 5).Return(active, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(assigned, nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, repo.ErrNoCandidate, err)
		}).Return(repo.ErrNoCandidate).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Reassign(ctx, prID, oldRev)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, repo.ErrNoCandidate, err)
	reviewerProv.AssertNotCalled(t, "ReassignReviewer", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPullRequestService_Reassign_ReassignReviewerError(t *testing.T) {
	ctx := context.Background()
	prID := "re-3"
	oldRev := "r1"

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	current := &models.PullRequest{ID: prID, Title: "Reassign err", AuthorId: "a1", Status: pr.StatusOpen}
	author := &models.User{ID: "a1", TeamID: 3}
	active := []string{"a1", "r1", "r2"}
	assigned := []string{"r1"}
	reassignErr := errors.New("reassign failed")

	prCtrl.On("GetById", ctx, prID).Return(current, nil).Once()
	userGetter.On("GetById", ctx, "a1").Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, 3).Return(active, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(assigned, nil).Once()
	reviewerProv.On("ReassignReviewer", ctx, prID, oldRev, mock.AnythingOfType("string")).Return(reassignErr).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, reassignErr, err)
		}).Return(reassignErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Reassign(ctx, prID, oldRev)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, reassignErr, err)
}

func TestPullRequestService_Reassign_GetByIdError(t *testing.T) {
	ctx := context.Background()
	prID := "re-4"
	oldRev := "r1"

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	getErr := errors.New("get pr failed")
	prCtrl.On("GetById", ctx, prID).Return((*models.PullRequest)(nil), getErr).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, getErr, err)
		}).Return(getErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Reassign(ctx, prID, oldRev)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, getErr, err)
}

func TestPullRequestService_Reassign_GetActiveUsersError(t *testing.T) {
	ctx := context.Background()
	prID := "re-5"
	oldRev := "r1"

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	current := &models.PullRequest{ID: prID, Title: "Err active", AuthorId: "a1", Status: pr.StatusOpen}
	author := &models.User{ID: "a1", TeamID: 13}
	actErr := errors.New("active users failed")

	prCtrl.On("GetById", ctx, prID).Return(current, nil).Once()
	userGetter.On("GetById", ctx, "a1").Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, 13).Return(([]string)(nil), actErr).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		err := fn(ctx)
		assert.Error(t, err)
		assert.Equal(t, actErr, err)
	}).Return(actErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Reassign(ctx, prID, oldRev)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, actErr, err)
}

func TestPullRequestService_Reassign_GetReviewersError(t *testing.T) {
	ctx := context.Background()
	prID := "re-6"
	oldRev := "r1"

	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)
	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	current := &models.PullRequest{ID: prID, Title: "Err reviewers", AuthorId: "a1", Status: pr.StatusOpen}
	author := &models.User{ID: "a1", TeamID: 1}
	revErr := errors.New("get reviewers failed")

	prCtrl.On("GetById", ctx, prID).Return(current, nil).Once()
	userGetter.On("GetById", ctx, "a1").Return(author, nil).Once()
	userGetter.On("GetActiveUsersIDInTeam", ctx, 1).Return([]string{"a1", "r1", "r2"}, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(([]string)(nil), revErr).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		err := fn(ctx)
		assert.Error(t, err)
		assert.Equal(t, revErr, err)
	}).Return(revErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Reassign(ctx, prID, oldRev)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, revErr, err)
}
