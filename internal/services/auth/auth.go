package auth

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/golang-jwt/jwt"
)

const (
	tokenTTL = 72 * time.Hour
)

type Service struct {
	secretKey string
	tokenTTL  time.Duration

	logger *logrus.Logger
}

func New(secretKey string, logger *logrus.Logger) *Service {
	return &Service{
		secretKey: secretKey,
		tokenTTL:  tokenTTL,
		logger:    logger,
	}
}

type tokenClaims struct {
	UserID int64 `json:"user_id"`
	jwt.StandardClaims
}

func (s *Service) GenerateToken(userID int64) (ss string, err error) {
	claims := tokenClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err = token.SignedString([]byte(s.secretKey))
	if err != nil {
		return ss, fmt.Errorf("failed to signed string: %w", err)
	}
	return
}

func (s *Service) ParseToken(accessToken string) (int64, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&tokenClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return []byte(s.secretKey), nil
		})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token with claims, %w", err)
	}
	if claims, ok := token.Claims.(*tokenClaims); ok && token.Valid {
		return claims.UserID, nil
	}
	return 0, errors.New("token claims are not of type *tokenClaims or not valid")
}

func (s *Service) GenerateUserSalt(complexity int64) string {
	digits := "0123456789"
	specials := "~=+%^*/()[]{}/!@#$?|"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials

	set := []byte(all)
	salt := make([]byte, complexity)
	for i := range salt {
		salt[i] = set[rand.Intn(len(set))]
	}
	return string(salt)
}

func (s *Service) GeneratePasswordHash(password string, salt string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}
