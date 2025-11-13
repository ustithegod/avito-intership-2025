package service

import (
	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/models"
	"context"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=PrProvider
type PrProvider interface {
	GetUserReviews(ctx context.Context, userID string) ([]*models.PullRequest, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TeamIDProvider
type TeamIDProvider interface {
	GetTeamNameByID(ctx context.Context, teamID int) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=UserChanger
type UserChanger interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	GetById(ctx context.Context, userID string) (*models.User, error)
}

type UserService struct {
	trm            TransactionManager
	prProvider     PrProvider
	userChanger    UserChanger
	teamIDProvider TeamIDProvider
}

func NewUserService(trm TransactionManager, prProvider PrProvider, userChanger UserChanger, teamIDProvider TeamIDProvider) *UserService {
	return &UserService{
		trm:            trm,
		prProvider:     prProvider,
		userChanger:    userChanger,
		teamIDProvider: teamIDProvider,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*api.UserSchema, error) {
	resp := &api.UserSchema{}

	err := s.trm.Do(ctx, func(ctx context.Context) error {
		err := s.userChanger.SetIsActive(ctx, userID, isActive)
		if err != nil {
			return err
		}

		user, err := s.userChanger.GetById(ctx, userID)
		if err != nil {
			return err
		}

		teamName, err := s.teamIDProvider.GetTeamNameByID(ctx, user.TeamID)
		if err != nil {
			return err
		}

		resp.UserID = user.ID
		resp.Username = user.Name
		resp.TeamName = teamName
		resp.IsActive = user.IsActive

		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *UserService) GetReview(ctx context.Context, userID string) (*api.GetReviewResponse, error) {
	prs, err := s.prProvider.GetUserReviews(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := &api.GetReviewResponse{
		UserID:       userID,
		PullRequests: make([]api.PullRequestShort, 0, len(prs)),
	}
	for _, pr := range prs {
		short := api.PullRequestShort{
			ID:       pr.ID,
			Name:     pr.Title,
			AuthorID: pr.AuthorId,
			Status:   pr.Status,
		}

		resp.PullRequests = append(resp.PullRequests, short)
	}

	return resp, err
}
