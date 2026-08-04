package auth

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
)

type Service struct {
	userRepository repository.UserRepository

	jwt *JWTManager

	config *config.Config
}

func NewService(
	userRepository repository.UserRepository,
	config *config.Config,
) *Service {

	return &Service{
		userRepository: userRepository,
		jwt: NewJWTManager(
			config.JWTSecret,
			config.JWTExpiration,
		),
		config: config,
	}
}

func (s *Service) Login(
	ctx context.Context,
	req LoginRequest,
) (*LoginResponse, error) {

	user, err := s.userRepository.GetByEmail(
		ctx,
		req.Email,
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	if !checkPasswordHash(
		req.Password,
		user.Password,
	) {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwt.GenerateToken(
		user.ID.String(),
		user.Email,
	)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: token,

		TokenType: "Bearer",

		ExpiresIn: s.config.JWTExpiration.String(),

		User: UserInfo{
			ID: user.ID.String(),

			FirstName: user.FirstName,

			LastName: user.LastName,

			Email: user.Email,

			IsActive: user.IsActive,
		},
	}, nil
}

func (s *Service) VerifyToken(
	token string,
) (*Claims, error) {

	claims, err := s.jwt.VerifyToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *Service) GetCurrentUser(
	ctx context.Context,
	userID string,
) (*UserInfo, error) {

	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := s.userRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		ID: user.ID.String(),

		FirstName: user.FirstName,

		LastName: user.LastName,

		Email: user.Email,

		IsActive: user.IsActive,
	}, nil
}

func checkPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
