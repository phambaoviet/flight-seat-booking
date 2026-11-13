package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService interface {
    Generate(userID string) (string, error)
    Verify(token string) (*Claims, error)
}
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTTokenService struct {
	secret    []byte
	expiresIn time.Duration
}	

func (s *JWTTokenService) Generate(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *JWTTokenService) Verify(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            if token.Method != jwt.SigningMethodHS256 {
                return nil, errors.New("unexpected signing method")
            }

            return []byte(s.secret), nil
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
func NewTokenService(secret []byte, expiresIn time.Duration) *JWTTokenService{
	return &JWTTokenService{
		secret:    secret,
		expiresIn: expiresIn,
	}
}