package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fenmo/expense-tracker/internal/model"
	"github.com/fenmo/expense-tracker/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, input model.RegisterInput) (*model.AuthResponse, error)
	Login(ctx context.Context, input model.LoginInput) (*model.AuthResponse, error)
	ValidateToken(tokenStr string) (uuid.UUID, error)
}

type authService struct {
	userRepo  repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{userRepo: userRepo, jwtSecret: []byte(jwtSecret)}
}

func (s *authService) Register(ctx context.Context, input model.RegisterInput) (*model.AuthResponse, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Email == "" {
		return nil, fmt.Errorf("%w: email is required", model.ErrInvalidInput)
	}
	if len(input.Password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", model.ErrInvalidInput)
	}

	existing, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: email already registered", model.ErrInvalidInput)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, input.Email, string(hash))
	if err != nil {
		return nil, err
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, User: user}, nil
}

func (s *authService) Login(ctx context.Context, input model.LoginInput) (*model.AuthResponse, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("%w: invalid email or password", model.ErrInvalidInput)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, fmt.Errorf("%w: invalid email or password", model.ErrInvalidInput)
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}
	return &model.AuthResponse{Token: token, User: user}, nil
}

func (s *authService) ValidateToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims")
	}
	sub, err := claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("missing subject")
	}
	id, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid subject")
	}
	return id, nil
}

func (s *authService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}
