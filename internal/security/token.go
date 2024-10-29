package security

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userId string) (string, error) {
	claims := &jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 2).Unix(),
		"iss": "devbook",
		"sub": userId,
	}

	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwt.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ValidateToken(header string) (*jwt.Token, error) {
	tokenString := extractToken(header)

	if tokenString == "" {
		return nil, errors.New("missing or invalid token")
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil

}

func extractToken(header string) string {
	h := strings.Split(header, " ")

	if len(h) != 2 || h[0] != "Bearer" {
		return ""
	}

	return h[1]
}
