package manager

import (
	"context"
	"errors"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
}

type UserClaims struct {
	jwt.StandardClaims
	UserId string `json:"user_id"`
}

type JWTToken struct {
	AccessToken string
}

func (jw *JWTToken) GetAccessToken() string {
	if jw == nil {
		return ""
	}
	return jw.AccessToken
}

func NewJWTManager(secretKey string, tokenDuration time.Duration, logger *logrus.Logger) *JWTManager {
	return &JWTManager{secretKey, tokenDuration}
}

func (manager *JWTManager) Generate(userID string) (*JWTToken, error) {
	var err error
	now := time.Now()

	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(manager.tokenDuration).Unix(),
			IssuedAt:  now.Unix(),
		},
		UserId: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(manager.secretKey))
	if err != nil {
		return nil, err
	}

	return &JWTToken{AccessToken: tokenString}, nil

}

func (manager *JWTManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.New("unexpected token signing method")
			}

			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, errors.New("invalid token: " + err.Error())
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (manager *JWTManager) GetAccessToken(ctx context.Context) (*string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.New("Unauthorized")
	}

	if _, ok := md["authorization"]; !ok {
		return nil, errors.New("Unauthorized")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return nil, errors.New("Unauthorized")

	}

	split := strings.Split(values[0], " ")
	accessToken := split[0]
	if len(split) > 1 {
		accessToken = split[1]
	}

	return &accessToken, nil
}
