package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var key []byte = []byte(os.Getenv("SIGNING_PRIVATE_KEY"))

type claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type authTokenContext struct{}

var tokenContext = authTokenContext{}

func (m claims) Validate() error {
	if m.UserID != m.Subject {
		return errors.New("user_id is not same as subject field")
	}
	return nil
}

type JwtToken struct {
	token *jwt.Token
}

func (j *JwtToken) ClaimsToContext(ctx context.Context) (context.Context, error) {
	retrievedClaims, ok := j.token.Claims.(*claims)
	if !ok {
		return ctx, errors.New("unable to parse claims")
	}
	fmt.Printf("\n\nWsadzam do kontekst: %v]\n\n", retrievedClaims)
	return context.WithValue(ctx, tokenContext, retrievedClaims), nil
}

func (j *JwtToken) Sign() (string, error) {
	rawToken, err := j.token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("creating new jwt token: %w", err)
	}
	return rawToken, nil
}

func ClaimsFromContext(ctx context.Context) (*claims, error) {
	retrievedClaims, ok := ctx.Value(tokenContext).(*claims)
	if !ok {
		return nil, errors.New("claims not in context")
	}

	return retrievedClaims, nil
}

func ParseJwtToken(token string) (*JwtToken, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (any, error) {
		return key, nil
	}, jwt.WithExpirationRequired(),
		jwt.WithIssuer("ecomm"),
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, fmt.Errorf("parsing jwt token: %v", err)
	}
	return &JwtToken{token: parsedToken}, nil
}

func NewJwtToken(userID string) *JwtToken {
	return &JwtToken{token: jwt.NewWithClaims(jwt.SigningMethodHS256, newJwtClaims("ecomm", userID))}
}

func newJwtClaims(issuer, subject string) claims {
	return claims{
		UserID: subject,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
			Subject:   subject,
			Audience:  []string{subject},
		},
	}
}
