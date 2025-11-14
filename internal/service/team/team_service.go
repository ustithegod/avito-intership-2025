package team

import (
	"context"

	"avito-intership-2025/internal/http/api"
	"avito-intership-2025/internal/models"
	"avito-intership-2025/internal/service"
)

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=TeamProvider
type TeamProvider interface {
	Create(ctx context.Context, teamName string) (int, error)
	GetByTeamName(ctx context.Context, teamName string) (*models.Team, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name=UserProvider
type UserProvider interface {
	Save(ctx context.Context, user *models.User) (string, error)
	GetUsersInTeam(ctx context.Context, teamName string) ([]*models.User, error)
}

type TeamService struct {
	teamProvider TeamProvider
	userProvider UserProvider
	trm          service.TransactionManager
}

func NewTeamService(
	trm service.TransactionManager,
	teamProvider TeamProvider,
	userProvider UserProvider,
) *TeamService {
	return &TeamService{
		teamProvider: teamProvider,
		userProvider: userProvider,
		trm:          trm,
	}
}

func (s *TeamService) Add(ctx context.Context, teamName string, users []api.TeamMember) (*api.TeamSchema, error) {
	resp := &api.TeamSchema{}
	members := make([]api.TeamMember, 0, len(users))

	err := s.trm.Do(ctx, func(ctx context.Context) error {
		teamID, err := s.teamProvider.Create(ctx, teamName)
		if err != nil {
			return err
		}

		for _, u := range users {
			user := &models.User{
				ID:       u.UserID,
				Name:     u.Username,
				TeamID:   teamID,
				IsActive: u.IsActive,
			}

			_, err := s.userProvider.Save(ctx, user)
			if err != nil {
				return err
			}

			member := api.TeamMember{
				UserID:   user.ID,
				Username: user.Name,
				IsActive: user.IsActive,
			}

			members = append(members, member)
		}

		resp.TeamName = teamName
		resp.Members = members

		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *TeamService) Get(ctx context.Context, teamName string) (*api.TeamSchema, error) {
	resp := &api.TeamSchema{}

	_, err := s.teamProvider.GetByTeamName(ctx, teamName)
	if err != nil {
		return nil, err
	}

	users, err := s.userProvider.GetUsersInTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	members := make([]api.TeamMember, 0, len(users))
	for _, u := range users {
		member := api.TeamMember{
			UserID:   u.ID,
			Username: u.Name,
			IsActive: u.IsActive,
		}

		members = append(members, member)
	}

	resp.TeamName = teamName
	resp.Members = members

	return resp, nil
}
