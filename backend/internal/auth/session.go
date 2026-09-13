package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var randomReader = rand.Reader

const (
	SessionCookieName = "x_clone_session"
	StateCookieName   = "x_clone_oauth_state"
	NonceCookieName   = "x_clone_oauth_nonce"
	PKCECookieName    = "x_clone_pkce_verifier"
)

type CookieConfig struct {
	Domain     string
	Secure     bool
	NamePrefix string
}

func (c CookieConfig) SessionName() string {
	if c.NamePrefix == "" {
		return SessionCookieName
	}
	return c.NamePrefix + "_session"
}

func (c CookieConfig) StateName() string {
	if c.NamePrefix == "" {
		return StateCookieName
	}
	return c.NamePrefix + "_oauth_state"
}

func (c CookieConfig) NonceName() string {
	if c.NamePrefix == "" {
		return NonceCookieName
	}
	return c.NamePrefix + "_oauth_nonce"
}

func (c CookieConfig) PKCEName() string {
	if c.NamePrefix == "" {
		return PKCECookieName
	}
	return c.NamePrefix + "_pkce_verifier"
}

func NewRandomToken() (string, error) {
	var value [32]byte
	if _, err := randomReader.Read(value[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value[:]), nil
}

func SignSession(userID, secret string, now time.Time) (string, error) {
	if userID == "" || secret == "" {
		return "", errors.New("session user and secret are required")
	}
	expiresAt := now.Add(24 * time.Hour).Unix()
	payload := fmt.Sprintf("%s.%d", userID, expiresAt)
	signature := sign(payload, secret)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + signature, nil
}

func VerifySession(value, secret string, now time.Time) (string, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 || secret == "" {
		return "", errors.New("invalid session")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", errors.New("invalid session")
	}
	payload := string(payloadBytes)
	if !hmac.Equal([]byte(sign(payload, secret)), []byte(parts[1])) {
		return "", errors.New("invalid session")
	}
	payloadParts := strings.Split(payload, ".")
	if len(payloadParts) != 2 {
		return "", errors.New("invalid session")
	}
	expiresAt, err := strconv.ParseInt(payloadParts[1], 10, 64)
	if err != nil || now.Unix() >= expiresAt {
		return "", errors.New("invalid session")
	}
	return payloadParts[0], nil
}

func SetSessionCookie(w http.ResponseWriter, value string, config CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.SessionName(),
		Value:    value,
		Path:     "/",
		Domain:   config.Domain,
		MaxAge:   int((24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   config.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter, config CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     config.SessionName(),
		Value:    "",
		Path:     "/",
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   config.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearLegacySessionCookie(w http.ResponseWriter, config CookieConfig) {
	if config.SessionName() == SessionCookieName {
		return
	}
	legacyConfig := config
	legacyConfig.NamePrefix = ""
	ClearSessionCookie(w, legacyConfig)
}

func SetTemporaryCookie(w http.ResponseWriter, name, value string, config CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   config.Domain,
		MaxAge:   int((10 * time.Minute).Seconds()),
		HttpOnly: true,
		Secure:   config.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearTemporaryCookie(w http.ResponseWriter, name string, config CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   config.Domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   config.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func sign(payload, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
