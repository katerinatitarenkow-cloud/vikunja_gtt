package googlecalendar

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strings"

	"code.vikunja.io/api/pkg/config"

	"golang.org/x/oauth2"
)

var scopes = []string{
	"https://www.googleapis.com/auth/calendar.app.created",
	"https://www.googleapis.com/auth/calendar.calendarlist.readonly",
	"https://www.googleapis.com/auth/calendar.events.readonly",
}

func IsConfigured() bool {
	return config.GoogleCalendarEnabled.GetBool() &&
		strings.TrimSpace(config.GoogleCalendarClientID.GetString()) != "" &&
		strings.TrimSpace(config.GoogleCalendarClientSecret.GetString()) != "" &&
		strings.TrimSpace(config.GoogleCalendarRedirectURL.GetString()) != ""
}

func OAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.GoogleCalendarClientID.GetString(),
		ClientSecret: config.GoogleCalendarClientSecret.GetString(),
		RedirectURL:  config.GoogleCalendarRedirectURL.GetString(),
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}
}

func NewState() (string, string, error) {
	raw := make([]byte, 32)

	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	state := base64.RawURLEncoding.EncodeToString(raw)

	return state, HashState(state), nil
}

func HashState(state string) string {
	sum := sha256.Sum256([]byte(state))
	return hex.EncodeToString(sum[:])
}

func AuthorizationURL(state string) string {
	return OAuthConfig().AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.SetAuthURLParam("include_granted_scopes", "true"),
	)
}

func Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if strings.TrimSpace(code) == "" {
		return nil, errors.New("google authorization code is empty")
	}

	return OAuthConfig().Exchange(ctx, code)
}

func encryptionKey() [32]byte {
	return sha256.Sum256(
		[]byte(config.ServiceSecret.GetString()),
	)
}

func EncryptRefreshToken(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", errors.New("google refresh token is empty")
	}

	key := encryptionKey()

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encrypted := gcm.Seal(nil, nonce, []byte(value), nil)

	payload := append(nonce, encrypted...)

	return base64.RawStdEncoding.EncodeToString(payload), nil
}

func DecryptRefreshToken(value string) (string, error) {
	payload, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}

	key := encryptionKey()

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(payload) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted google refresh token")
	}

	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}
