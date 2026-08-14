package auth

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

type Service struct {
	userRepository repository.UserRepository
	jwt            *JWTManager
	config         *config.Config
	log            *zap.Logger
}

func NewService(
	userRepository repository.UserRepository,
	jwtManager *JWTManager,
	config *config.Config,
	log *zap.Logger,
) *Service {

	return &Service{
		userRepository: userRepository,
		jwt:            jwtManager,
		config:         config,
		log:            log,
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
		logger.WarnContext(
			ctx,
			s.log,
			"user login failed",
			zap.String("email", req.Email),
		)

		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		logger.WarnContext(
			ctx,
			s.log,
			"login attempt for inactive user",
			zap.String("user_id", user.ID.String()),
		)

		return nil, ErrInactiveUser
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(req.Password),
	) != nil {
		logger.WarnContext(
			ctx,
			s.log,
			"user login failed",
			zap.String("user_id", user.ID.String()),
		)

		return nil, ErrInvalidCredentials
	}

	token, err := s.jwt.GenerateToken(
		user.ID.String(),
		user.Email,
	)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"user login successful",
		zap.String("user_id", user.ID.String()),
	)

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

	return s.jwt.VerifyToken(token)
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
