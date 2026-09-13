package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/apperrors"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

type AuthConfig struct {
	Issuer        string
	AuthorizeURL  string
	TokenURL      string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	PostLoginURL  string
	Cookie        auth.CookieConfig
	SessionSecret string
}

type AuthController struct {
	config     AuthConfig
	httpClient *http.Client
	users      UserService
	now        func() time.Time
}

type UserService interface {
	FindOrCreateByOIDC(ctx context.Context, subject, email, displayName string) (models.User, error)
	FindByID(ctx context.Context, id string) (models.User, error)
}

func NewAuthController(config AuthConfig, users UserService) *AuthController {
	return &AuthController{
		config:     config,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		users:      users,
		now:        time.Now,
	}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)

	state, err := auth.NewRandomToken()
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.DependencyUnavailable.Wrap(err, "authentication is temporarily unavailable"))
		return
	}
	nonce, err := auth.NewRandomToken()
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.DependencyUnavailable.Wrap(err, "authentication is temporarily unavailable"))
		return
	}
	verifier, err := auth.NewRandomToken()
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.DependencyUnavailable.Wrap(err, "authentication is temporarily unavailable"))
		return
	}

	auth.SetTemporaryCookie(w, auth.StateCookieName, state, c.config.Cookie)
	auth.SetTemporaryCookie(w, auth.NonceCookieName, nonce, c.config.Cookie)
	auth.SetTemporaryCookie(w, auth.PKCECookieName, verifier, c.config.Cookie)

	values := url.Values{}
	values.Set("client_id", c.config.ClientID)
	values.Set("response_type", "code")
	values.Set("scope", "openid email profile")
	values.Set("redirect_uri", c.config.RedirectURL)
	values.Set("state", state)
	values.Set("nonce", nonce)
	values.Set("code_challenge", auth.CodeChallenge(verifier))
	values.Set("code_challenge_method", "S256")

	http.Redirect(w, r, c.config.AuthorizeURL+"?"+values.Encode(), http.StatusFound)
}

func (c *AuthController) Callback(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	stateCookie := cookieValue(r, auth.StateCookieName)
	if state == "" || code == "" || stateCookie != state {
		apperrors.ErrorHandler(w, r, apperrors.BadParam.Wrap(
			fmt.Errorf("callback validation failed: state=%t code=%t state_cookie=%t state_match=%t", state != "", code != "", stateCookie != "", stateCookie == state),
			"invalid authentication callback",
		))
		return
	}
	nonce := cookieValue(r, auth.NonceCookieName)
	verifier := cookieValue(r, auth.PKCECookieName)
	if nonce == "" || verifier == "" {
		apperrors.ErrorHandler(w, r, apperrors.BadParam.Wrap(
			fmt.Errorf("callback temporary cookie missing: nonce=%t verifier=%t", nonce != "", verifier != ""),
			"invalid authentication callback",
		))
		return
	}

	token, err := c.exchangeCode(r.Context(), code, verifier)
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.DependencyUnavailable.Wrap(err, "authentication is temporarily unavailable"))
		return
	}
	claims, err := auth.VerifyIDToken(r.Context(), c.httpClient, token.IDToken, c.config.Issuer, c.config.ClientID, nonce, c.now())
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.BadParam.Wrap(err, "invalid authentication token"))
		return
	}
	user, err := c.users.FindOrCreateByOIDC(r.Context(), claims.Subject, claims.Email, claims.Name)
	if err != nil {
		apperrors.ErrorHandler(w, r, err)
		return
	}
	session, err := auth.SignSession(user.ID, c.config.SessionSecret, c.now())
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.DependencyUnavailable.Wrap(err, "authentication is temporarily unavailable"))
		return
	}
	auth.SetSessionCookie(w, session, c.config.Cookie)
	auth.ClearTemporaryCookie(w, auth.StateCookieName, c.config.Cookie)
	auth.ClearTemporaryCookie(w, auth.NonceCookieName, c.config.Cookie)
	auth.ClearTemporaryCookie(w, auth.PKCECookieName, c.config.Cookie)
	http.Redirect(w, r, c.config.PostLoginURL, http.StatusFound)
}

func (c *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)

	userID, err := auth.UserID(r.Context())
	if err != nil {
		apperrors.ErrorHandler(w, r, apperrors.Unauthorized.Wrap(err, "login is required"))
		return
	}
	user, err := c.users.FindByID(r.Context(), userID)
	if err != nil {
		apperrors.ErrorHandler(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		apperrors.ErrorHandler(w, r, err)
	}
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	setAuthNoStore(w)
	auth.ClearSessionCookie(w, c.config.Cookie)
	w.WriteHeader(http.StatusNoContent)
}

func (c *AuthController) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.SessionCookieName)
		if err != nil {
			apperrors.ErrorHandler(w, r, apperrors.Unauthorized.Wrap(err, "login is required"))
			return
		}
		userID, err := auth.VerifySession(cookie.Value, c.config.SessionSecret, c.now())
		if err != nil {
			apperrors.ErrorHandler(w, r, apperrors.Unauthorized.Wrap(err, "login is required"))
			return
		}
		next.ServeHTTP(w, r.WithContext(auth.WithUserID(r.Context(), userID)))
	})
}

func (c *AuthController) exchangeCode(ctx context.Context, code, verifier string) (auth.TokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("client_id", c.config.ClientID)
	values.Set("client_secret", c.config.ClientSecret)
	values.Set("code", code)
	values.Set("redirect_uri", c.config.RedirectURL)
	values.Set("code_verifier", verifier)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.TokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return auth.TokenResponse{}, errors.New("cannot prepare token request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return auth.TokenResponse{}, errors.New("cannot exchange authorization code")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return auth.TokenResponse{}, errors.New("cannot exchange authorization code")
	}
	var token auth.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil || token.IDToken == "" {
		return auth.TokenResponse{}, errors.New("cannot decode token response")
	}
	return token, nil
}

func cookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func setAuthNoStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
}
