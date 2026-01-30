package auth

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
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

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
	IdToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
}

var httpClient http.Client = http.Client{Timeout: 10 * time.Second}

func (a *Authenticator) RefreshToken(ctx context.Context, actualRefreshToken string) (RefreshTokenResponse, error) {
	tokenUrl := a.oauth.Endpoint.TokenURL
	parsedUrl, err := url.Parse("https://" + tokenUrl)
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("parsing refresh token url: %w", err)
	}
	payload := strings.NewReader("grant_type=refresh_token&client_id={yourClientId}&client_secret=%7ByourClientSecret%7D&refresh_token=%7ByourRefreshToken%7D")
	params := url.Values{}
	params.Add("grant_type", "refresh_token")
	params.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	params.Add("refresh_token", actualRefreshToken)
	parsedUrl.RawQuery = params.Encode()

	req, _ := http.NewRequestWithContext(ctx, "POST", tokenUrl, payload)

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("requesting new token: %w", err)
	}
	defer res.Body.Close()

	var refreshTokenResp RefreshTokenResponse
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("reading response body: %w", err)
	}
	err = json.Unmarshal(body, &refreshTokenResp)
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("unmarshalling response body: %w", err)
	}
	return refreshTokenResp, nil
}

func (a *Authenticator) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return a.oauth.AuthCodeURL(state, opts...)
}

func (a *Authenticator) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return a.oauth.Exchange(ctx, code, opts...)
}

func VerifySession(session *session.Session) (isExpired bool, err error) {
	expiry, ok := session.Get("expiry").(int64)
	if !ok {
		return false, errors.New("expiry field not set")
	}
	now := time.Now().Add(24 * time.Hour).Unix()
	isExpired = now > expiry

	return isExpired, nil
}
