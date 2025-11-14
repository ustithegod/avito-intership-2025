package pr

import (
	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/models"
	repo "avito-intership-2025/internal/repository"
	"avito-intership-2025/internal/service"
	"context"
	"errors"
	"math/rand/v2"
)

const (
	StatusOpen   = "OPEN"
	StatusMerged = "MERGED"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=PrController
type PrController interface {
	Create(ctx context.Context, pr *models.PullRequest) (string, error)
	GetById(ctx context.Context, prID string) (*models.PullRequest, error)
	MarkAsMerged(ctx context.Context, prID string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=ReviewerProvider
type ReviewerProvider interface {
	GetPrReviewers(ctx context.Context, prID string) ([]string, error)
	AssignReviewer(ctx context.Context, prID, userID string) error
	ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error
	DeleteReviewer(ctx context.Context, prID, userID string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=UserGetter
type UserGetter interface {
	GetActiveUsersIDInTeam(ctx context.Context, teamID int) ([]string, error)
	GetById(ctx context.Context, userID string) (*models.User, error)
}

type PullRequestService struct {
	prController     PrController
	userGetter       UserGetter
	reviewerProvider ReviewerProvider
	trm              service.TransactionManager
}

func NewPullRequestService(
	trm service.TransactionManager,
	prController PrController,
	reviewerProvider ReviewerProvider,
	userGetter UserGetter,
) *PullRequestService {
	return &PullRequestService{
		trm:              trm,
		prController:     prController,
		userGetter:       userGetter,
		reviewerProvider: reviewerProvider,
	}
}

var (
	ErrTryMergeMerged = errors.New("cannot reassign on merged PR")
)

func (s *PullRequestService) Create(ctx context.Context, prID, prName, authorId string) (*api.PullRequestSchema, error) {

	pr := &models.PullRequest{
		ID:       prID,
		Title:    prName,
		AuthorId: authorId,
		Status:   StatusOpen,
	}

	resp := &api.PullRequestSchema{
		AssignedReviewers: make([]string, 0, 2),
	}

	err := s.trm.Do(ctx, func(ctx context.Context) error {
		createdPrID, err := s.prController.Create(ctx, pr)
		if err != nil {
			return err
		}

		author, err := s.userGetter.GetById(ctx, authorId)
		if err != nil {
			return err
		}

		teamID := author.TeamID
		activeUsers, err := s.userGetter.GetActiveUsersIDInTeam(ctx, teamID)
		if err != nil {
			return err
		}

		reviewers := getRandomUsers(activeUsers, 2, authorId)

		for _, r := range reviewers {
			err = s.reviewerProvider.AssignReviewer(ctx, createdPrID, r)
			if err != nil {
				return err
			}
		}

		toPullRequestSchema(resp, pr, reviewers)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *PullRequestService) Merge(ctx context.Context, prID string) (*api.PullRequestSchema, error) {

	resp := &api.PullRequestSchema{
		AssignedReviewers: make([]string, 0, 2),
	}

	err := s.trm.Do(ctx, func(ctx context.Context) error {
		pr, err := s.prController.GetById(ctx, prID)
		if err != nil {
			return err
		}

		if pr.Status == StatusOpen {
			err = s.prController.MarkAsMerged(ctx, pr.ID)
		}

		pr, err = s.prController.GetById(ctx, prID)
		if err != nil {
			return err
		}

		reviewers, err := s.reviewerProvider.GetPrReviewers(ctx, prID)
		if err != nil {
			return err
		}

		toPullRequestSchema(resp, pr, reviewers)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *PullRequestService) Reassign(ctx context.Context, prID, oldRev string) (*api.ReassignResponse, error) {
	resp := &api.ReassignResponse{
		PullRequest: api.PullRequestSchema{
			AssignedReviewers: make([]string, 0, 2),
		},
	}

	err := s.trm.Do(ctx, func(ctx context.Context) error {
		pr, err := s.prController.GetById(ctx, prID)
		if err != nil {
			return err
		}

		if pr.Status == StatusMerged {
			return repo.ErrPRMerged
		}

		author, err := s.userGetter.GetById(ctx, pr.AuthorId)
		if err != nil {
			return err
		}
		teamID := author.TeamID
		activeUsers, err := s.userGetter.GetActiveUsersIDInTeam(ctx, teamID)
		if err != nil {
			return err
		}

		assignedReviewers, err := s.reviewerProvider.GetPrReviewers(ctx, prID)
		if err != nil {
			return err
		}

		exludedReviewers := []string{author.ID}
		exludedReviewers = append(exludedReviewers, assignedReviewers...)

		var newRev string
		if len(exludedReviewers) >= len(activeUsers) {
			return repo.ErrNoCandidate
		} else {
			newRev = getRandomUsers(activeUsers, 1, exludedReviewers...)[0]

			err = s.reviewerProvider.ReassignReviewer(ctx, prID, oldRev, newRev)
		}
		if err != nil {
			return err
		}

		pr, err = s.prController.GetById(ctx, prID)
		if err != nil {
			return err
		}

		reviewers, err := s.reviewerProvider.GetPrReviewers(ctx, prID)
		if err != nil {
			return err
		}

		toPullRequestSchema(&resp.PullRequest, pr, reviewers)
		resp.ReplacedBy = newRev
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func toPullRequestSchema(resp *api.PullRequestSchema, pr *models.PullRequest, reviewers []string) {
	resp.ID = pr.ID
	resp.Name = pr.Title
	resp.AuthorID = pr.AuthorId
	resp.Status = pr.Status
	resp.AssignedReviewers = append(resp.AssignedReviewers, reviewers...)
	resp.MergedAt = pr.MergedAt
}

func getRandomUsers(candidates []string, maxCount int, excludedIDs ...string) []string {
	if len(candidates) == 0 || maxCount <= 0 {
		return []string{}
	}

	excluded := make(map[string]struct{}, len(excludedIDs))
	for _, id := range excludedIDs {
		excluded[id] = struct{}{}
	}

	var available []string
	for _, user := range candidates {
		if _, skip := excluded[user]; !skip {
			available = append(available, user)
		}
	}

	if len(available) == 0 {
		return []string{}
	}

	rand.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})

	if len(available) > maxCount {
		return available[:maxCount]
	}
	return available
}
