package auth

import (
	"context"
	"errors"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"golang.org/x/oauth2"
)

func UserIDFromRequest(c fiber.Ctx) string {
	userID, ok := c.Locals("user_id").(string)
	if ok && userID != "" {
		return userID
	}

	sess := session.FromContext(c)
	if sess == nil || sess.Session == nil {
		return ""
	}
	userID, ok = sess.Get("user_id").(string)
	if !ok || userID == "" {
		return ""
	}
	c.Locals("user_id", userID)

	return userID
}

func PersistUserID(c fiber.Ctx, userID string) {
	if userID == "" {
		return
	}

	c.Locals("user_id", userID)

	sess := session.FromContext(c)
	if sess == nil || sess.Session == nil {
		return
	}
	sess.Set("user_id", userID)
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
