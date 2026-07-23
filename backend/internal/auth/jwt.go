package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const expiracao = 24 * time.Hour

type Claims struct {
	UID      string `json:"uid"`
	Email    string `json:"email"`
	Nome     string `json:"nome"`
	IsAdmin  bool   `json:"isAdmin"`
	Aprovado bool   `json:"aprovado"`
	jwt.RegisteredClaims
}

func gerarToken(secret string, c Claims) (string, error) {
	c.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiracao)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(secret))
}

func parseToken(secret, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inesperado: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token inválido: %w", err)
	}
	return claims, nil
}
