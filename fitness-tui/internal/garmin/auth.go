package garmin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sstent/go-garth"
	"gopkg.in/yaml.v3"
)

const sessionTimeout = 30 * time.Minute

type Auth struct {
	client      *garth.Client
	username    string
	password    string
	sessionPath string
}

type Session struct {
	Cookies   []*http.Cookie `yaml:"cookies"`
	ExpiresAt time.Time      `yaml:"expires_at"`
}

func NewAuth(username, password, storagePath string) *Auth {
	return &Auth{
		client:      garth.New(),
		username:    username,
		password:    password,
		sessionPath: filepath.Join(storagePath, "garmin_session.yaml"),
	}
}

func (a *Auth) Connect(ctx context.Context) error {
	if sess, err := a.loadSession(); err == nil {
		if time.Now().Before(sess.ExpiresAt) {
			a.client.SetCookies(sess.Cookies)
			return nil
		}
	}
	return a.login(ctx)
}

func (a *Auth) login(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, sessionTimeout)
	defer cancel()

	for attempt := 1; attempt <= 3; attempt++ {
		if err := a.client.Login(ctx, a.username, a.password); err == nil {
			return a.saveSession()
		}
		time.Sleep(time.Duration(attempt*attempt) * time.Second)
	}
	return errors.New("authentication failed after 3 attempts")
}

func (a *Auth) saveSession() error {
	sess := Session{
		Cookies:   a.client.Cookies(),
		ExpiresAt: time.Now().Add(sessionTimeout),
	}

	data, err := yaml.Marshal(sess)
	if err != nil {
		return fmt.Errorf("session marshal failed: %w", err)
	}

	return os.WriteFile(a.sessionPath, data, 0600)
}

func (a *Auth) loadSession() (*Session, error) {
	data, err := os.ReadFile(a.sessionPath)
	if err != nil {
		return nil, err
	}

	var sess Session
	if err := yaml.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("session parse failed: %w", err)
	}
	return &sess, nil
}
