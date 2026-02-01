package auth

import (
	"context"
	"encoding/gob"
	"errors"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2/middleware/session"
	"golang.org/x/oauth2"
)

type Profile struct {
	Name string `json:"name"`
}

func init() {
	gob.Register(Profile{})
}

type userIDContextType struct{}
type sessionToContext struct{}

var userIDContext = userIDContextType{}
var sessionContext = sessionToContext{}

func UserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContext, userID)
}

func SessionToContext(ctx context.Context, session *session.Session) context.Context {
	return context.WithValue(ctx, sessionContext, session)
}

// panics when session not in context
func SessionFromContext(ctx context.Context) *session.Session {
	session, ok := ctx.Value(sessionContext).(*session.Session)
	if !ok {
		panic("unable to retrieve session from context")
	}
	return session
}

// panics when userID not in context
func UserIDFromContext(ctx context.Context) string {
	retrievedClaims, ok := ctx.Value(userIDContext).(string)
	if !ok {
		panic("unable to retrieve userID from context")
	}

	return retrievedClaims
}

// Authenticator is used to authenticate our users.
type Authenticator struct {
	oidc  *oidc.Provider
	oauth oauth2.Config
}

// New instantiates the *Authenticator.
func New() (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+os.Getenv("AUTH0_DOMAIN")+"/",
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("AUTH0_CALLBACK_URL"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", oidc.ScopeOfflineAccess},
	}

	return &Authenticator{
		oidc:  provider,
		oauth: conf,
	}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.oauth.ClientID,
	}

	return a.oidc.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func (a *Authenticator) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return a.oauth.AuthCodeURL(state, opts...)
}

func (a *Authenticator) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return a.oauth.Exchange(ctx, code, opts...)
}

const TokenExpiryTime = 3 * 24 * time.Hour

func VerifySession(session *session.Session) (isExpired bool, err error) {
	expiry, ok := session.Get("expiry").(int64)
	if !ok {
		return false, errors.New("expiry field not set")
	}
	now := time.Now().Add(24 * time.Hour).Unix()
	isExpired = now > expiry

	return isExpired, nil
}
