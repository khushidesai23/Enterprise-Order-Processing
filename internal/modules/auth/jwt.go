package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey []byte
	expiry    time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`

	Email string `json:"email"`

	jwt.RegisteredClaims
}

func NewJWTManager(
	secret string,
	expiry time.Duration,
) *JWTManager {

	return &JWTManager{
		secretKey: []byte(secret),
		expiry:    expiry,
	}
}

func (j *JWTManager) GenerateToken(
	userID string,
	email string,
) (string, error) {

	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userID,

			IssuedAt: jwt.NewNumericDate(now),

			ExpiresAt: jwt.NewNumericDate(
				now.Add(j.expiry),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(j.secretKey)
}

func (j *JWTManager) VerifyToken(
	tokenString string,
) (*Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {

			if token.Method != jwt.SigningMethodHS256 {

				return nil, errors.New("unexpected signing method")
			}

			return j.secretKey, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {

		return nil, errors.New("invalid token")
	}

	return claims, nil
}