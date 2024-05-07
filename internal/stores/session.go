package stores

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

const sessionCookieName = "recipes.session"

type CookieSessionStore struct {
	s    *securecookie.SecureCookie
	opts CookieSessionOptions
}

type CookieSessionOptions struct {
	HashKey       []byte
	EncryptionKey []byte
	Secure        bool
	Duration      time.Duration
}

func NewCookieSessionStore(opts CookieSessionOptions) CookieSessionStore {
	return CookieSessionStore{
		securecookie.New(opts.HashKey, opts.EncryptionKey),
		opts,
	}
}

func (c *CookieSessionStore) Create(_ context.Context, w http.ResponseWriter, id string) error {
	data := map[string]any{
		"id": id,
	}

	encoded, err := c.s.Encode(sessionCookieName, data)
	if err != nil {
		return fmt.Errorf("failed to encode session data: %w", err)
	}

	cookie := &http.Cookie{
		Name:   sessionCookieName,
		Value:  encoded,
		Path:   "/",
		Secure: c.opts.Secure,
		MaxAge: int(c.opts.Duration.Seconds()),
	}
	http.SetCookie(w, cookie)

	return nil
}

func (c *CookieSessionStore) IsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}

	var data map[string]any
	if err := c.s.Decode(sessionCookieName, cookie.Value, &data); err != nil {
		return false
	}

	rawID, ok := data["id"]
	if !ok {
		return false
	}

	id, ok := rawID.(string)
	if !ok {
		return false
	}

	return id != ""
}

func (c *CookieSessionStore) UserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", fmt.Errorf("could not find session cookie: %w", err)
	}

	var data map[string]any
	if err := c.s.Decode(sessionCookieName, cookie.Value, &data); err != nil {
		return "", fmt.Errorf("failed to decode session cookie: %w", err)
	}

	rawID, ok := data["id"]
	if !ok {
		return "", errors.New("missing 'id' in session cookie")
	}

	id, ok := rawID.(string)
	if !ok {
		return "", fmt.Errorf("could not convert id %v to a string", rawID)
	}

	return id, err
}
