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

	autherrors "github.com/accelolabs/avito-tamagochi/backend/internal/auth/errors"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
	"github.com/google/uuid"
)

type fakeService struct {
	register func(model.RegisterInput) (*model.User, *model.Session, error)
	login    func(model.LoginInput) (*model.User, *model.Session, error)
	logout   func(uuid.UUID) error
}

func (f fakeService) Register(_ context.Context, input model.RegisterInput) (*model.User, *model.Session, error) {
	return f.register(input)
}
func (f fakeService) Login(_ context.Context, input model.LoginInput) (*model.User, *model.Session, error) {
	if f.login == nil {
		return nil, nil, errors.New("not implemented")
	}
	return f.login(input)
}
func (f fakeService) Logout(_ context.Context, sessionID uuid.UUID) error {
	if f.logout == nil {
		return nil
	}
	return f.logout(sessionID)
}
func (fakeService) ValidateSession(context.Context, uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, sql.ErrNoRows
}
func (fakeService) FindUser(context.Context, uuid.UUID) (*model.User, error) {
	return nil, sql.ErrNoRows
}

func TestRegisterSetsSessionCookieAndReturnsUser(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	service := fakeService{register: func(input model.RegisterInput) (*model.User, *model.Session, error) {
		if input.Email != " User@Example.com " || input.Password != "password" || input.DisplayName != "Player" {
			t.Fatalf("unexpected register input: %+v", input)
		}
		return &model.User{ID: userID, Email: "user@example.com", DisplayName: "Player", CreatedAt: time.Unix(10, 0).UTC()}, &model.Session{ID: sessionID, ExpiresAt: time.Now().Add(time.Hour)}, nil
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
	handler := New(fakeService{register: func(model.RegisterInput) (*model.User, *model.Session, error) {
		return nil, nil, autherrors.ErrEmailAlreadyExists
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

func TestRegisterRejectsUnexpectedJSONField(t *testing.T) {
	handler := New(fakeService{register: func(model.RegisterInput) (*model.User, *model.Session, error) {
		t.Fatal("service must not be called for an invalid request")
		return nil, nil, nil
	}})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{"email":"user@example.com","password":"password","displayName":"Player","admin":"true"}`))
	response := httptest.NewRecorder()

	handler.register(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestLoginSetsSessionCookieAndReturnsUser(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	service := fakeService{login: func(input model.LoginInput) (*model.User, *model.Session, error) {
		if input.Email != "user@example.com" || input.Password != "password" {
			t.Fatalf("unexpected login input: %+v", input)
		}
		return &model.User{ID: userID, Email: input.Email, DisplayName: "Player"}, &model.Session{ID: sessionID, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}}
	handler := New(service)
	mux := http.NewServeMux()
	handler.SetRoutes(mux)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"user@example.com","password":"password"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Name != sessionCookieName || cookie.Value != sessionID.String() || !cookie.HttpOnly {
		t.Fatalf("unexpected session cookie: %+v", cookie)
	}
}

func TestLogoutClearsCookieWithoutRequestCookie(t *testing.T) {
	handler := New(fakeService{})
	mux := http.NewServeMux()
	handler.SetRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	cookie := response.Result().Cookies()[0]
	if cookie.Name != sessionCookieName || cookie.MaxAge != -1 {
		t.Fatalf("unexpected cleared cookie: %+v", cookie)
	}
}
