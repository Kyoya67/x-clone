package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"time"
)

type TokenResponse struct {
	IDToken string `json:"id_token"`
}

type IDTokenClaims struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Issuer        string `json:"iss"`
	Audience      string `json:"aud"`
	ExpiresAt     int64  `json:"exp"`
	Nonce         string `json:"nonce"`
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Algorithm string `json:"alg"`
	Use       string `json:"use"`
	N         string `json:"n"`
	E         string `json:"e"`
}

func VerifyIDToken(ctx context.Context, client *http.Client, token, issuer, audience, nonce string, now time.Time) (IDTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return IDTokenClaims{}, errors.New("invalid id token")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return IDTokenClaims{}, errors.New("invalid id token")
	}
	var header struct {
		KeyID     string `json:"kid"`
		Algorithm string `json:"alg"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil || header.Algorithm != "RS256" {
		return IDTokenClaims{}, errors.New("invalid id token")
	}
	key, err := fetchPublicKey(ctx, client, issuer, header.KeyID)
	if err != nil {
		return IDTokenClaims{}, err
	}
	signed := parts[0] + "." + parts[1]
	digest := sha256.Sum256([]byte(signed))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return IDTokenClaims{}, errors.New("invalid id token")
	}
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return IDTokenClaims{}, errors.New("invalid id token")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return IDTokenClaims{}, errors.New("invalid id token")
	}
	var claims IDTokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return IDTokenClaims{}, errors.New("invalid id token")
	}
	if claims.Issuer != issuer || claims.Audience != audience || claims.Nonce != nonce || claims.ExpiresAt <= now.Unix() || claims.Subject == "" {
		return IDTokenClaims{}, errors.New("invalid id token")
	}
	return claims, nil
}

func fetchPublicKey(ctx context.Context, client *http.Client, issuer, keyID string) (*rsa.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(issuer, "/")+"/.well-known/jwks.json", nil)
	if err != nil {
		return nil, errors.New("cannot prepare jwks request")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("cannot fetch jwks")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("cannot fetch jwks")
	}
	var keys jwks
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, errors.New("cannot decode jwks")
	}
	for _, key := range keys.Keys {
		if key.KeyID == keyID && key.KeyType == "RSA" {
			return rsaPublicKey(key)
		}
	}
	return nil, errors.New("cannot find jwks key")
}

func rsaPublicKey(key jwk) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, errors.New("invalid jwks key")
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, errors.New("invalid jwks key")
	}
	e := 0
	for _, b := range eBytes {
		e = e*256 + int(b)
	}
	if e == 0 {
		return nil, errors.New("invalid jwks key")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: e}, nil
}
