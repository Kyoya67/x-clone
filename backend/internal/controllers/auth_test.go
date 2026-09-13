package controllers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/auth"
	"github.com/Kaminashi-Inc/ENG-1103_Kyoya67/backend/internal/models"
)

const testAuthSessionSecret = "test-auth-session-secret"

type fakeAuthUserService struct {
	user models.User
	err  error
}

func (s fakeAuthUserService) FindOrCreateByOIDC(context.Context, string, string, string) (models.User, error) {
	return s.user, s.err
}

func (s fakeAuthUserService) FindByID(context.Context, string) (models.User, error) {
	return s.user, s.err
}

func (s fakeAuthUserService) UpdateProfile(context.Context, string, models.UpdateUserProfileRequest) (models.User, error) {
	return s.user, s.err
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestAuthLoginDisablesCache(t *testing.T) {
	controller := NewAuthController(AuthConfig{
		AuthorizeURL: "https://example.com/oauth2/authorize",
		ClientID:     "client-id",
		RedirectURL:  "https://example.com/auth/callback",
	}, fakeAuthUserService{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/login", nil)

	controller.Login(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, response.Code)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store cache control, got %q", response.Header().Get("Cache-Control"))
	}
	if !strings.HasPrefix(response.Header().Get("Location"), "https://example.com/oauth2/authorize?") {
		t.Fatalf("unexpected redirect location: %s", response.Header().Get("Location"))
	}
}

func TestAuthLoginConvertsTokenGenerationFailure(t *testing.T) {
	tests := []struct {
		name      string
		failAfter int
	}{
		{name: "state", failAfter: 0},
		{name: "nonce", failAfter: 1},
		{name: "verifier", failAfter: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewAuthController(AuthConfig{}, fakeAuthUserService{})
			calls := 0
			controller.randomToken = func() (string, error) {
				if calls == tt.failAfter {
					return "", errors.New("random source failed")
				}
				calls++
				return "token", nil
			}

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/auth/login", nil)

			controller.Login(response, request)

			if response.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
			}
			if strings.Contains(response.Body.String(), "random source failed") {
				t.Fatalf("internal random failure leaked: %s", response.Body.String())
			}
		})
	}
}

func TestAuthCallbackRejectsInvalidStateWithoutLeakingDetails(t *testing.T) {
	controller := NewAuthController(AuthConfig{}, fakeAuthUserService{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/callback?state=query-state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: "x_clone_oauth_state", Value: "cookie-state"})

	controller.Callback(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store cache control, got %q", response.Header().Get("Cache-Control"))
	}
	if !strings.Contains(response.Body.String(), `"errCode":"R001"`) {
		t.Fatalf("expected bad parameter error body, got %s", response.Body.String())
	}
	if strings.Contains(response.Body.String(), "state_match") {
		t.Fatalf("internal callback details leaked: %s", response.Body.String())
	}
}

func TestAuthCallbackRejectsMissingTemporaryCookie(t *testing.T) {
	controller := NewAuthController(AuthConfig{}, fakeAuthUserService{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/callback?state=state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: auth.StateCookieName, Value: "state"})

	controller.Callback(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if strings.Contains(response.Body.String(), "verifier") || strings.Contains(response.Body.String(), "nonce") {
		t.Fatalf("internal cookie details leaked: %s", response.Body.String())
	}
}

func TestAuthCallbackConvertsTokenExchangeFailure(t *testing.T) {
	controller := NewAuthController(AuthConfig{
		TokenURL: "https://example.com/oauth2/token",
	}, fakeAuthUserService{})
	controller.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network details")
	})}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/callback?state=state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: auth.StateCookieName, Value: "state"})
	request.AddCookie(&http.Cookie{Name: auth.NonceCookieName, Value: "nonce"})
	request.AddCookie(&http.Cookie{Name: auth.PKCECookieName, Value: "verifier"})

	controller.Callback(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}
	if !strings.Contains(response.Body.String(), `"errCode":"D001"`) {
		t.Fatalf("expected dependency error body, got %s", response.Body.String())
	}
	if strings.Contains(response.Body.String(), "network details") {
		t.Fatalf("internal token exchange details leaked: %s", response.Body.String())
	}
}

func TestAuthCallbackRejectsInvalidIDToken(t *testing.T) {
	controller := NewAuthController(AuthConfig{
		TokenURL: "https://example.com/oauth2/token",
	}, fakeAuthUserService{})
	controller.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if strings.HasSuffix(request.URL.Path, "/oauth2/token") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"id_token":"invalid"}`)),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"keys":[]}`)),
			Header:     make(http.Header),
		}, nil
	})}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/callback?state=state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: auth.StateCookieName, Value: "state"})
	request.AddCookie(&http.Cookie{Name: auth.NonceCookieName, Value: "nonce"})
	request.AddCookie(&http.Cookie{Name: auth.PKCECookieName, Value: "verifier"})

	controller.Callback(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
}

func TestAuthCallbackSetsSessionAndRedirects(t *testing.T) {
	user := models.User{ID: "user-1"}
	controller := NewAuthController(AuthConfig{
		TokenURL:      "https://example.com/oauth2/token",
		PostLoginURL:  "https://example.com/",
		SessionSecret: testAuthSessionSecret,
	}, fakeAuthUserService{user: user})
	controller.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"id_token":"id-token"}`)),
			Header:     make(http.Header),
		}, nil
	})}
	controller.verifyIDToken = func(context.Context, *http.Client, string, string, string, string, time.Time) (auth.IDTokenClaims, error) {
		return auth.IDTokenClaims{Subject: "subject", Email: "user@example.com", Name: "User"}, nil
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/callback?state=state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: auth.StateCookieName, Value: "state"})
	request.AddCookie(&http.Cookie{Name: auth.NonceCookieName, Value: "nonce"})
	request.AddCookie(&http.Cookie{Name: auth.PKCECookieName, Value: "verifier"})

	controller.Callback(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, response.Code)
	}
	if response.Header().Get("Location") != "https://example.com/" {
		t.Fatalf("unexpected redirect location: %s", response.Header().Get("Location"))
	}
	hasSessionCookie := false
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == auth.SessionCookieName && cookie.Value != "" {
			hasSessionCookie = true
		}
	}
	if !hasSessionCookie {
		t.Fatalf("expected session cookie, got %#v", response.Result().Cookies())
	}
}

func TestAuthCallbackReturnsUserServiceError(t *testing.T) {
	controller := NewAuthController(AuthConfig{
		TokenURL: "https://example.com/oauth2/token",
	}, fakeAuthUserService{err: errors.New("secret database details")})
	controller.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"id_token":"id-token"}`)),
			Header:     make(http.Header),
		}, nil
	})}
	controller.verifyIDToken = func(context.Context, *http.Client, string, string, string, string, time.Time) (auth.IDTokenClaims, error) {
		return auth.IDTokenClaims{Subject: "subject", Email: "user@example.com", Name: "User"}, nil
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/callback?state=state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: auth.StateCookieName, Value: "state"})
	request.AddCookie(&http.Cookie{Name: auth.NonceCookieName, Value: "nonce"})
	request.AddCookie(&http.Cookie{Name: auth.PKCECookieName, Value: "verifier"})

	controller.Callback(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
	if strings.Contains(response.Body.String(), "secret database details") {
		t.Fatalf("internal user service error leaked: %s", response.Body.String())
	}
}

func TestAuthCallbackConvertsSessionSignFailure(t *testing.T) {
	controller := NewAuthController(AuthConfig{
		TokenURL: "https://example.com/oauth2/token",
	}, fakeAuthUserService{})
	controller.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"id_token":"invalid"}`)),
			Header:     make(http.Header),
		}, nil
	})}
	controller.signSession = func(string, string, time.Time) (string, error) {
		return "", errors.New("sign failed")
	}
	controller.verifyIDToken = func(context.Context, *http.Client, string, string, string, string, time.Time) (auth.IDTokenClaims, error) {
		return auth.IDTokenClaims{Subject: "subject", Email: "user@example.com", Name: "User"}, nil
	}

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/callback?state=state&code=code", nil)
	request.AddCookie(&http.Cookie{Name: auth.StateCookieName, Value: "state"})
	request.AddCookie(&http.Cookie{Name: auth.NonceCookieName, Value: "nonce"})
	request.AddCookie(&http.Cookie{Name: auth.PKCECookieName, Value: "verifier"})

	controller.Callback(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}
}

func TestAuthMeReturnsCurrentUser(t *testing.T) {
	user := models.User{ID: "user-1", Handle: "handle", DisplayName: "display"}
	controller := NewAuthController(AuthConfig{}, fakeAuthUserService{user: user})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	request = request.WithContext(auth.WithUserID(request.Context(), user.ID))

	controller.Me(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !strings.Contains(response.Body.String(), `"id":"user-1"`) {
		t.Fatalf("expected user body, got %s", response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store cache control, got %q", response.Header().Get("Cache-Control"))
	}
}

func TestAuthMeReturnsServiceError(t *testing.T) {
	controller := NewAuthController(AuthConfig{}, fakeAuthUserService{err: errors.New("secret database details")})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	request = request.WithContext(auth.WithUserID(request.Context(), "user-1"))

	controller.Me(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.Code)
	}
	if strings.Contains(response.Body.String(), "secret database details") {
		t.Fatalf("internal service error leaked: %s", response.Body.String())
	}
}

func TestAuthMeRequiresLogin(t *testing.T) {
	controller := NewAuthController(AuthConfig{}, fakeAuthUserService{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)

	controller.Me(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthLogoutClearsSessionCookie(t *testing.T) {
	controller := NewAuthController(AuthConfig{}, fakeAuthUserService{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

	controller.Logout(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != auth.SessionCookieName || cookies[0].MaxAge != -1 {
		t.Fatalf("expected cleared session cookie, got %#v", cookies)
	}
}

func TestAuthMiddlewareRequiresSessionCookie(t *testing.T) {
	controller := NewAuthController(AuthConfig{SessionSecret: testAuthSessionSecret}, fakeAuthUserService{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	controller.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	})).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthMiddlewareRejectsInvalidSessionCookie(t *testing.T) {
	controller := NewAuthController(AuthConfig{SessionSecret: testAuthSessionSecret}, fakeAuthUserService{})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: "invalid"})
	controller.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler should not be called")
	})).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, response.Code)
	}
}

func TestAuthMiddlewarePassesUserID(t *testing.T) {
	now := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	session, err := auth.SignSession("user-1", testAuthSessionSecret, now)
	if err != nil {
		t.Fatal(err)
	}
	controller := NewAuthController(AuthConfig{SessionSecret: testAuthSessionSecret}, fakeAuthUserService{})
	controller.now = func() time.Time { return now }

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: session})

	controller.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.UserID(r.Context())
		if err != nil || userID != "user-1" {
			t.Fatalf("expected user ID in context, got %q, err=%v", userID, err)
		}
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}
}

func TestAuthExchangeCodeReturnsToken(t *testing.T) {
	controller := NewAuthController(AuthConfig{
		TokenURL:     "https://example.com/oauth2/token",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURL:  "https://example.com/auth/callback",
	}, fakeAuthUserService{})
	controller.httpClient = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatal(err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if values.Get("code") != "code" || values.Get("code_verifier") != "verifier" {
			t.Fatalf("unexpected token request body: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"id_token":"id-token"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	token, err := controller.exchangeCode(context.Background(), "code", "verifier")
	if err != nil {
		t.Fatal(err)
	}
	if token.IDToken != "id-token" {
		t.Fatalf("unexpected token: %+v", token)
	}
}

func TestAuthExchangeCodeRejectsNonOKResponse(t *testing.T) {
	controller := NewAuthController(AuthConfig{
		TokenURL: "https://example.com/oauth2/token",
	}, fakeAuthUserService{})
	controller.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewBufferString(`{"error":"invalid_grant"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := controller.exchangeCode(context.Background(), "code", "verifier")
	if err == nil || !strings.Contains(err.Error(), "status=400") {
		t.Fatalf("expected non-OK error, got %v", err)
	}
}

func TestAuthExchangeCodeRejectsInvalidTokenResponse(t *testing.T) {
	controller := NewAuthController(AuthConfig{
		TokenURL: "https://example.com/oauth2/token",
	}, fakeAuthUserService{})
	controller.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := controller.exchangeCode(context.Background(), "code", "verifier")
	if err == nil || !strings.Contains(err.Error(), "cannot decode token response") {
		t.Fatalf("expected decode error, got %v", err)
	}
}
