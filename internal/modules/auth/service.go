package auth

import (
	"context"
	"errors"

	"rmp-api/pkg/hash"
	pkgjwt "rmp-api/pkg/jwt"
)

type Service struct {
	repo      *Repository
	jwtSecret string
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	if !hash.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}

	accessToken, expiresIn, err := pkgjwt.GenerateAccessToken(user.ID, user.Role, user.BranchID, s.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	refreshToken, err := pkgjwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User: UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      user.Role,
			BranchID:  user.BranchID,
		},
	}, nil
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*LoginResponse, error) {
	claims, err := pkgjwt.ValidateToken(req.RefreshToken, s.jwtSecret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	userID, _ := claims["sub"].(string)
	if userID == "" {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	accessToken, expiresIn, err := pkgjwt.GenerateAccessToken(user.ID, user.Role, user.BranchID, s.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	refreshToken, err := pkgjwt.GenerateRefreshToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User: UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      user.Role,
			BranchID:  user.BranchID,
		},
	}, nil
}
