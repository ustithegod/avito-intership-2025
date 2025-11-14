package pr_test

import (
	"avito-intership-2025/internal/models"
	"avito-intership-2025/internal/service/mocks"
	"avito-intership-2025/internal/service/pr"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullRequestService_Merge_Success_EmptyReviewers(t *testing.T) {
	ctx := context.Background()
	prID := "merge-empty-reviewers"

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	now := time.Now()
	merged := &models.PullRequest{
		ID:       prID,
		Title:    "Already merged",
		AuthorId: "author-1",
		Status:   pr.StatusMerged,
		MergedAt: &now,
	}

	// First GetById shows merged, so MarkAsMerged must not be called
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	// The service asks again after the conditional
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return([]string{}, nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		assert.NoError(t, fn(ctx))
	}).Return(nil).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, prID, resp.ID)
	assert.Equal(t, "Already merged", resp.Name)
	assert.Equal(t, pr.StatusMerged, resp.Status)
	assert.NotNil(t, resp.MergedAt)
	assert.Empty(t, resp.AssignedReviewers)

	// Ensure MarkAsMerged wasn't called
	prCtrl.AssertNotCalled(t, "MarkAsMerged", ctx, prID)
}

func TestPullRequestService_Merge_TRMReturnsError_AfterInnerOK(t *testing.T) {
	ctx := context.Background()
	prID := "merge-trm-error"

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	now := time.Now()
	merged := &models.PullRequest{
		ID:       prID,
		Title:    "PR",
		AuthorId: "u1",
		Status:   pr.StatusMerged,
		MergedAt: &now,
	}
	reviewers := []string{"r1", "r2"}

	// All inner calls succeed
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(reviewers, nil).Once()

	trmErr := errors.New("transaction failed")
	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		// inner execution OK
		assert.NoError(t, fn(ctx))
	}).Return(trmErr).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, trmErr, err)
}

func TestPullRequestService_Merge_MarkAsMergedError_PrStillOpen(t *testing.T) {
	ctx := context.Background()
	prID := "merge-mark-error-open"

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	open := &models.PullRequest{ID: prID, Title: "PR", AuthorId: "u1", Status: pr.StatusOpen}
	// After failed merge, the PR still reported as OPEN
	stillOpen := &models.PullRequest{ID: prID, Title: "PR", AuthorId: "u1", Status: pr.StatusOpen}
	reviewers := []string{"r1"}

	prCtrl.On("GetById", ctx, prID).Return(open, nil).Once()
	prCtrl.On("MarkAsMerged", ctx, prID).Return(errors.New("merge failed")).Once()
	prCtrl.On("GetById", ctx, prID).Return(stillOpen, nil).Once()
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(reviewers, nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		// Service ignores MarkAsMerged error and continues
		assert.NoError(t, fn(ctx))
	}).Return(nil).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	// Status remains OPEN, because implementation uses second GetById result as-is
	assert.Equal(t, pr.StatusOpen, resp.Status)
	assert.Equal(t, reviewers, resp.AssignedReviewers)
}

func TestPullRequestService_Merge_MarkAsMergedError_SecondGetByIdError(t *testing.T) {
	ctx := context.Background()
	prID := "merge-mark-error-second-get"

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	open := &models.PullRequest{ID: prID, Title: "PR", AuthorId: "u1", Status: pr.StatusOpen}
	secondErr := errors.New("second get failed")

	prCtrl.On("GetById", ctx, prID).Return(open, nil).Once()
	prCtrl.On("MarkAsMerged", ctx, prID).Return(errors.New("merge failed")).Once()
	prCtrl.On("GetById", ctx, prID).Return((*models.PullRequest)(nil), secondErr).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
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

func TestPullRequestService_Merge_Success_NilReviewersSlice(t *testing.T) {
	ctx := context.Background()
	prID := "merge-nil-reviewers"

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	prCtrl := mocks.NewPrController(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	merged := &models.PullRequest{ID: prID, Title: "PR", AuthorId: "u1", Status: pr.StatusMerged}
	// Both GetById calls return merged
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	prCtrl.On("GetById", ctx, prID).Return(merged, nil).Once()
	// reviewerProv returns nil slice without error
	reviewerProv.On("GetPrReviewers", ctx, prID).Return(([]string)(nil), nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).Run(func(args mock.Arguments) {
		fn := args.Get(1).(func(context.Context) error)
		assert.NoError(t, fn(ctx))
	}).Return(nil).Once()

	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, nil)
	resp, err := svc.Merge(ctx, prID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, pr.StatusMerged, resp.Status)
	assert.Empty(t, resp.AssignedReviewers)
}
