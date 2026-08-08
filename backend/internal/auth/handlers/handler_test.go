package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
	"github.com/google/uuid"
)

type fakeService struct {
	register func(auth.RegisterInput) (*auth.User, *auth.Session, error)
}

func (f fakeService) Register(_ context.Context, input auth.RegisterInput) (*auth.User, *auth.Session, error) {
	return f.register(input)
}
func (fakeService) Login(context.Context, auth.LoginInput) (*auth.User, *auth.Session, error) {
	return nil, nil, errors.New("not implemented")
}
func (fakeService) Logout(context.Context, uuid.UUID) error { return nil }
func (fakeService) ValidateSession(context.Context, uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, sql.ErrNoRows
}
func (fakeService) FindUser(context.Context, uuid.UUID) (*auth.User, error) {
	return nil, sql.ErrNoRows
}

func TestRegisterSetsSessionCookieAndReturnsUser(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	service := fakeService{register: func(input auth.RegisterInput) (*auth.User, *auth.Session, error) {
		if input.Email != " User@Example.com " || input.Password != "password" || input.DisplayName != "Player" {
			t.Fatalf("unexpected register input: %+v", input)
		}
		return &auth.User{ID: userID, Email: "user@example.com", DisplayName: "Player", CreatedAt: time.Unix(10, 0).UTC()}, &auth.Session{ID: sessionID, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}}
	handler := New(service)
	mux := http.NewServeMux()
	handler.SetRoutes(mux)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":" User@Example.com ","password":"password","displayName":"Player"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Name != sessionCookieName || cookie.Value != sessionID.String() || !cookie.HttpOnly {
		t.Fatalf("unexpected session cookie: %+v", cookie)
	}
	if !strings.Contains(response.Body.String(), `"displayName":"Player"`) {
		t.Fatalf("response body does not contain user: %s", response.Body.String())
	}
}

func TestRegisterMapsDuplicateEmail(t *testing.T) {
	handler := New(fakeService{register: func(auth.RegisterInput) (*auth.User, *auth.Session, error) {
		return nil, nil, auth.ErrEmailAlreadyExists
	}})
	mux := http.NewServeMux()
	handler.SetRoutes(mux)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"user@example.com","password":"password","displayName":"Player"}`))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
	}
	if !strings.Contains(response.Body.String(), `"code":"email_already_exists"`) {
		t.Fatalf("unexpected error body: %s", response.Body.String())
	}
}
