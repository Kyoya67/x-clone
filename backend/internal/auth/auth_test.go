package auth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("random failed")
}

func TestUserIDContext(t *testing.T) {
	ctx := WithUserID(context.Background(), "user-1")

	userID, err := UserID(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if userID != "user-1" {
		t.Fatalf("expected user ID, got %q", userID)
	}
}

func TestUserIDRejectsMissingValue(t *testing.T) {
	if _, err := UserID(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	if _, err := UserID(WithUserID(context.Background(), "")); err == nil {
		t.Fatal("expected error")
	}
}

func TestCodeChallenge(t *testing.T) {
	got := CodeChallenge("verifier")
	want := "iMnq5o6zALKXGivsnlom_0F5_WYda32GHkxlV7mq7hQ"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestNewRandomToken(t *testing.T) {
	token, err := NewRandomToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
}

func TestNewRandomTokenReturnsRandomReaderError(t *testing.T) {
	original := randomReader
	randomReader = errorReader{}
	t.Cleanup(func() { randomReader = original })

	if _, err := NewRandomToken(); err == nil {
		t.Fatal("expected error")
	}
}

func TestSessionLifecycle(t *testing.T) {
	now := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	session, err := SignSession("user-1", "secret", now)
	if err != nil {
		t.Fatal(err)
	}

	userID, err := VerifySession(session, "secret", now)
	if err != nil {
		t.Fatal(err)
	}
	if userID != "user-1" {
		t.Fatalf("expected user ID, got %q", userID)
	}
}

func TestSignSessionRejectsMissingValue(t *testing.T) {
	if _, err := SignSession("", "secret", time.Now()); err == nil {
		t.Fatal("expected error")
	}
	if _, err := SignSession("user-1", "", time.Now()); err == nil {
		t.Fatal("expected error")
	}
}

func TestVerifySessionRejectsInvalidValue(t *testing.T) {
	now := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	session, err := SignSession("user-1", "secret", now)
	if err != nil {
		t.Fatal(err)
	}
	malformedPayload := base64.RawURLEncoding.EncodeToString([]byte("user-1"))
	malformedPayloadSession := malformedPayload + "." + sign("user-1", "secret")
	invalidExpiryPayload := "user-1.invalid"
	invalidExpirySession := base64.RawURLEncoding.EncodeToString([]byte(invalidExpiryPayload)) + "." + sign(invalidExpiryPayload, "secret")

	tests := []struct {
		name   string
		value  string
		secret string
		now    time.Time
	}{
		{name: "empty", value: "", secret: "secret", now: now},
		{name: "missing secret", value: session, secret: "", now: now},
		{name: "invalid base64", value: "xxx.signature", secret: "secret", now: now},
		{name: "invalid signature", value: session + "x", secret: "secret", now: now},
		{name: "malformed payload", value: malformedPayloadSession, secret: "secret", now: now},
		{name: "invalid expiry", value: invalidExpirySession, secret: "secret", now: now},
		{name: "expired", value: session, secret: "secret", now: now.Add(25 * time.Hour)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := VerifySession(tt.value, tt.secret, tt.now); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestSessionCookies(t *testing.T) {
	config := CookieConfig{Domain: "example.com", Secure: true, NamePrefix: "x_clone_stg"}
	recorder := httptest.NewRecorder()

	SetSessionCookie(recorder, "session", config)
	SetTemporaryCookie(recorder, config.StateName(), "state", config)
	ClearTemporaryCookie(recorder, config.StateName(), config)
	ClearSessionCookie(recorder, config)
	ClearLegacySessionCookie(recorder, config)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 5 {
		t.Fatalf("expected cookies, got %#v", cookies)
	}
	if cookies[0].Name != "x_clone_stg_session" || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].Domain != "example.com" {
		t.Fatalf("unexpected session cookie: %#v", cookies[0])
	}
	if cookies[3].Name != "x_clone_stg_session" || cookies[3].MaxAge != -1 {
		t.Fatalf("expected cleared session cookie, got %#v", cookies[3])
	}
	if cookies[4].Name != SessionCookieName || cookies[4].MaxAge != -1 {
		t.Fatalf("expected cleared legacy session cookie, got %#v", cookies[4])
	}
}

func TestCookieConfigUsesLegacyNamesWithoutPrefix(t *testing.T) {
	config := CookieConfig{}

	if config.SessionName() != SessionCookieName || config.StateName() != StateCookieName || config.NonceName() != NonceCookieName || config.PKCEName() != PKCECookieName {
		t.Fatalf("unexpected legacy names: %#v", config)
	}
}

func TestVerifyIDToken(t *testing.T) {
	now := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	key := newTestRSAKey(t)
	issuer := "https://issuer.example.com"
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/.well-known/jwks.json" {
			t.Fatalf("unexpected jwks path: %s", r.URL.Path)
		}
		body, err := json.Marshal(jwks{Keys: []jwk{testJWK(key.PublicKey)}})
		if err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: ioNopCloser(string(body)), Header: make(http.Header)}, nil
	})}
	token := signedTestJWT(t, key, IDTokenClaims{
		Subject:   "subject",
		Email:     "user@example.com",
		Name:      "User",
		Issuer:    issuer,
		Audience:  "client-id",
		ExpiresAt: now.Add(time.Hour).Unix(),
		Nonce:     "nonce",
	})

	claims, err := VerifyIDToken(context.Background(), client, token, issuer, "client-id", "nonce", now)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "subject" || claims.Email != "user@example.com" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifyIDTokenRejectsInvalidValues(t *testing.T) {
	now := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	key := newTestRSAKey(t)
	issuer := "https://issuer.example.com"
	body, err := json.Marshal(jwks{Keys: []jwk{testJWK(key.PublicKey)}})
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: ioNopCloser(string(body)), Header: make(http.Header)}, nil
	})}
	valid := signedTestJWT(t, key, IDTokenClaims{
		Subject:   "subject",
		Issuer:    issuer,
		Audience:  "client-id",
		ExpiresAt: now.Add(time.Hour).Unix(),
		Nonce:     "nonce",
	})
	unsignedInvalidPayload := signedTestJWTWithRawPayload(t, key, "%%%")
	unsignedInvalidJSON := signedTestJWTWithRawPayload(t, key, base64.RawURLEncoding.EncodeToString([]byte("{")))

	tests := []struct {
		name  string
		token string
	}{
		{name: "invalid token format", token: "invalid"},
		{name: "invalid header base64", token: "%%%.payload.signature"},
		{name: "invalid header JSON", token: base64.RawURLEncoding.EncodeToString([]byte("{")) + ".payload.signature"},
		{name: "unsupported alg", token: base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","kid":"kid"}`)) + ".payload.signature"},
		{name: "invalid signature base64", token: strings.Join(strings.Split(valid, ".")[:2], ".") + ".%%%"},
		{name: "invalid payload base64", token: unsignedInvalidPayload},
		{name: "invalid payload JSON", token: unsignedInvalidJSON},
		{name: "wrong claims", token: signedTestJWT(t, key, IDTokenClaims{
			Subject:   "subject",
			Issuer:    issuer,
			Audience:  "client-id",
			ExpiresAt: now.Add(time.Hour).Unix(),
			Nonce:     "wrong",
		})},
		{name: "invalid signature", token: strings.Join(strings.Split(valid, ".")[:2], ".") + "." + base64.RawURLEncoding.EncodeToString([]byte("invalid-signature"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := VerifyIDToken(context.Background(), client, tt.token, issuer, "client-id", "nonce", now); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestVerifyIDTokenReturnsFetchPublicKeyError(t *testing.T) {
	now := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	key := newTestRSAKey(t)
	token := signedTestJWT(t, key, IDTokenClaims{
		Subject:   "subject",
		Issuer:    "https://issuer.example.com",
		Audience:  "client-id",
		ExpiresAt: now.Add(time.Hour).Unix(),
		Nonce:     "nonce",
	})
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusInternalServerError, Body: ioNopCloser(""), Header: make(http.Header)}, nil
	})}

	if _, err := VerifyIDToken(context.Background(), client, token, "https://issuer.example.com", "client-id", "nonce", now); err == nil {
		t.Fatal("expected error")
	}
}

func TestRSAPublicKeyRejectsInvalidJWK(t *testing.T) {
	tests := []struct {
		name string
		key  jwk
	}{
		{name: "invalid modulus", key: jwk{N: "%%%", E: base64.RawURLEncoding.EncodeToString([]byte{1})}},
		{name: "invalid exponent", key: jwk{N: base64.RawURLEncoding.EncodeToString([]byte{1}), E: "%%%"}},
		{name: "zero exponent", key: jwk{N: base64.RawURLEncoding.EncodeToString([]byte{1}), E: base64.RawURLEncoding.EncodeToString([]byte{0})}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := rsaPublicKey(tt.key); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func newTestRSAKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func testJWK(key rsa.PublicKey) jwk {
	return jwk{
		KeyID:     "kid",
		KeyType:   "RSA",
		Algorithm: "RS256",
		Use:       "sig",
		N:         base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
		E:         base64.RawURLEncoding.EncodeToString([]byte{byte(key.E >> 16), byte(key.E >> 8), byte(key.E)}),
	}
}

func signedTestJWT(t *testing.T, key *rsa.PrivateKey, claims IDTokenClaims) string {
	t.Helper()
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	return signedTestJWTWithRawPayload(t, key, base64.RawURLEncoding.EncodeToString(claimsJSON))
}

func signedTestJWTWithRawPayload(t *testing.T, key *rsa.PrivateKey, payload string) string {
	t.Helper()
	header := map[string]string{"alg": "RS256", "kid": "kid"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatal(err)
	}
	unsigned := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + payload
	digest := sha256.Sum256([]byte(unsigned))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature)
}

func TestFetchPublicKeyRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "non OK", status: http.StatusInternalServerError, body: ""},
		{name: "invalid JSON", status: http.StatusOK, body: "{"},
		{name: "missing key", status: http.StatusOK, body: `{"keys":[]}`},
		{name: "invalid key", status: http.StatusOK, body: `{"keys":[{"kid":"kid","kty":"RSA","n":"%%%","e":"AQAB"}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tt.status, Body: ioNopCloser(tt.body), Header: make(http.Header)}, nil
			})}

			if _, err := fetchPublicKey(context.Background(), client, "https://issuer.example.com", "kid"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestFetchPublicKeyRejectsInvalidIssuerURL(t *testing.T) {
	client := &http.Client{}
	if _, err := fetchPublicKey(context.Background(), client, string([]byte{0x7f}), "kid"); err == nil {
		t.Fatal("expected error")
	}
}

func ioNopCloser(body string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(body))
}
