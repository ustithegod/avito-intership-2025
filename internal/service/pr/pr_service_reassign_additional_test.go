package pr_test

import (
	"context"
	"testing"

	"avito-intership-2025/internal/models"
	repo "avito-intership-2025/internal/repository"
	"avito-intership-2025/internal/service/mocks"
	"avito-intership-2025/internal/service/pr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPullRequestService_Reassign_ReturnsErrPRMerged_WhenPRIsMerged(t *testing.T) {
	ctx := context.Background()
	prID := "pr-merged"
	oldRev := "user-old"

	// Mocks
	prCtrl := mocks.NewPrController(t)
	userGetter := mocks.NewUserGetter(t)
	reviewerProv := mocks.NewReviewerProvider(t)

	trm := &mocks.MockManager{}
	trm.Test(t)
	t.Cleanup(func() { trm.AssertExpectations(t) })

	// PR is already merged -> service should return repo.ErrPRMerged
	mergedPR := &models.PullRequest{
		ID:       prID,
		Title:    "Already merged PR",
		AuthorId: "author-1",
		Status:   pr.StatusMerged,
	}
	prCtrl.On("GetById", ctx, prID).Return(mergedPR, nil).Once()

	trm.On("Do", ctx, mock.AnythingOfType("func(context.Context) error")).
		Run(func(args mock.Arguments) {
			fn := args.Get(1).(func(context.Context) error)
			err := fn(ctx)
			assert.Error(t, err)
			assert.Equal(t, repo.ErrPRMerged, err)
		}).
		Return(repo.ErrPRMerged).
		Once()

	// SUT
	svc := pr.NewPullRequestService(trm, prCtrl, reviewerProv, userGetter)
	resp, err := svc.Reassign(ctx, prID, oldRev)

	// Assert
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Equal(t, repo.ErrPRMerged, err)

	// Ensure no further interactions happened
	userGetter.AssertNotCalled(t, "GetById", mock.Anything, mock.Anything)
	reviewerProv.AssertNotCalled(t, "GetPrReviewers", mock.Anything, mock.Anything)
	reviewerProv.AssertNotCalled(t, "ReassignReviewer", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
