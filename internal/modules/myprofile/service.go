package myprofile

import (
	"context"
	"errors"
	"strings"

	"clinic-api/internal/models"
	"clinic-api/pkg/hash"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpdateMyProfile(ctx context.Context, userID string, req UpdateMyProfileRequest) (*models.User, error) {
	user, err := s.repo.FindMyProfileByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.FirstName != nil {
		firstName := strings.TrimSpace(*req.FirstName)
		if firstName == "" {
			return nil, errors.New("first_name cannot be empty")
		}
		user.FirstName = firstName
	}

	if req.LastName != nil {
		lastName := strings.TrimSpace(*req.LastName)
		if lastName == "" {
			return nil, errors.New("last_name cannot be empty")
		}
		user.LastName = lastName
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		if email == "" {
			return nil, errors.New("email cannot be empty")
		}
		user.Email = email
	}

	if req.Phone != nil {
		phone := strings.TrimSpace(*req.Phone)
		if phone == "" {
			user.Phone = nil
		} else {
			user.Phone = &phone
		}
	}

	if req.AvatarURL != nil {
		avatarURL := strings.TrimSpace(*req.AvatarURL)
		if avatarURL == "" {
			user.AvatarURL = nil
		} else {
			user.AvatarURL = &avatarURL
		}
	}

	if err := s.repo.UpdateMyProfile(ctx, userID, user); err != nil {
		return nil, err
	}

	return s.repo.FindMyProfileByID(ctx, userID)
}

func (s *Service) ChangeMyPassword(ctx context.Context, userID string, req ChangeMyPasswordRequest) error {
	user, err := s.repo.FindMyProfileByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if req.OldPassword == nil {
		return errors.New("old_password is required")
	}

	if req.NewPassword == nil {
		return errors.New("new_password is required")
	}

	oldPassword := strings.TrimSpace(*req.OldPassword)
	if oldPassword == "" {
		return errors.New("old_password cannot be empty")
	}

	newPassword := strings.TrimSpace(*req.NewPassword)
	if newPassword == "" {
		return errors.New("new_password cannot be empty")
	}

	if !hash.CheckPassword(oldPassword, user.PasswordHash) {
		return errors.New("old_password is incorrect")
	}

	passwordHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	return s.repo.UpdateMyPassword(ctx, userID, passwordHash)
}

func (s *Service) GetMenusByRole(ctx context.Context, role string) ([]models.Menu, error) {
	flatMenus, err := s.repo.FindMenusByRole(ctx, role)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(flatMenus), nil
}

func buildMenuTree(flatMenus []models.Menu) []models.Menu {
	menuMap := make(map[string]*models.Menu)
	var roots []models.Menu

	for i := range flatMenus {
		flatMenus[i].Children = []models.Menu{}
		menuMap[flatMenus[i].ID] = &flatMenus[i]
	}

	for i := range flatMenus {
		if flatMenus[i].ParentID != nil && *flatMenus[i].ParentID != "" {
			if parent, ok := menuMap[*flatMenus[i].ParentID]; ok {
				parent.Children = append(parent.Children, flatMenus[i])
			}
		} else {
			roots = append(roots, flatMenus[i])
		}
	}

	for i := range roots {
		if mapped, ok := menuMap[roots[i].ID]; ok {
			roots[i].Children = mapped.Children
		}
	}

	return roots
}
