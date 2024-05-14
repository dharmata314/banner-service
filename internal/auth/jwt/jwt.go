package jwt

import (
	errMsg "banner-serivce/internal/api/err"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

type JWTManager struct {
	secret []byte
	log    *slog.Logger
}

func NewJWTManager(secret string, log *slog.Logger) *JWTManager {
	return &JWTManager{secret: []byte(secret), log: log}
}

func (manager *JWTManager) GenerateToken(username, role string, expiration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"usename": username,
		"role":    role,
		"exp":     time.Now().Add(expiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(manager.secret)
	if err != nil {
		manager.log.Error("Failed to sign token")
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (manager *JWTManager) VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			manager.log.Error("Unexpected signing method")
			return nil, errors.New("unexpected signing method")
		}
		return manager.secret, nil
	})
	if err != nil {
		manager.log.Error("failed to parse token", errMsg.Err(err))
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		manager.log.Error("invalid token")
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (manager *JWTManager) ExtractRoleFromToken(tokenString string) (string, error) {
	claims, err := manager.VerifyToken(tokenString)
	if err != nil {
		manager.log.Error("Error with token", errMsg.Err(err))
		return "", err
	}

	role, ok := claims["role"].(string)
	if !ok {
		manager.log.Error("Role not found in claims")
		return "", errors.New("role not found in token")
	}

	return role, nil
}

func (manager *JWTManager) ExtractRoleAndUsernameFromToken(tokenString string) (string, string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			manager.log.Error("Unexpected signing method")
			return nil, errors.New("unexpected signing method")
		}
		return manager.secret, nil
	})

	if err != nil {
		manager.log.Error("Failed to parse token", errMsg.Err(err))
		return "", "", fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		manager.log.Error("Invalid token")
		return "", "", errors.New("invalid token")
	}

	return claims.Username, claims.Role, nil
}
