package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	authmodel "github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
	"github.com/google/uuid"
)

type fakeNotificationService struct {
	result *notificationmodel.DispatchResult
	err    error
}

func (f fakeNotificationService) DispatchEnergy(context.Context, uuid.UUID) (*notificationmodel.DispatchResult, error) {
	return f.result, f.err
}

type fakeAuthService struct {
	userID uuid.UUID
}

func (f fakeAuthService) Register(context.Context, authmodel.RegisterInput) (*authmodel.User, *authmodel.Session, error) {
	return nil, nil, nil
}
func (f fakeAuthService) Login(context.Context, authmodel.LoginInput) (*authmodel.User, *authmodel.Session, error) {
	return nil, nil, nil
}
func (f fakeAuthService) Logout(context.Context, uuid.UUID) error { return nil }
func (f fakeAuthService) ValidateSession(context.Context, uuid.UUID) (uuid.UUID, error) {
	return f.userID, nil
}
func (f fakeAuthService) FindUser(context.Context, uuid.UUID) (*authmodel.User, error) {
	return nil, nil
}

func TestEnergyNotificationRouteRequiresAuthentication(t *testing.T) {
	mux := http.NewServeMux()
	New(fakeNotificationService{}, fakeAuthService{userID: uuid.New()}).SetRoutes(mux)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/mock-notifications/energy", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if response.Body.String() != "{\"code\":\"unauthorized\",\"message\":\"authentication is required\"}\n" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestEnergyNotificationResponses(t *testing.T) {
	threshold := 25
	tests := []struct {
		name   string
		result *notificationmodel.DispatchResult
		body   string
	}{
		{"sent", &notificationmodel.DispatchResult{Status: notificationmodel.StatusSent, Energy: 25, Threshold: &threshold}, "{\"status\":\"sent\",\"energy\":25,\"threshold\":25}\n"},
		{"skipped", &notificationmodel.DispatchResult{Status: notificationmodel.StatusSkipped, Energy: 76}, "{\"status\":\"skipped\",\"energy\":76,\"threshold\":null}\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := serveAuthenticated(t, fakeNotificationService{result: test.result})
			if response.Code != http.StatusOK || response.Body.String() != test.body {
				t.Fatalf("status = %d body = %q", response.Code, response.Body.String())
			}
			if got := response.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q", got)
			}
		})
	}
}

func TestEnergyNotificationMapsInternalErrorsSafely(t *testing.T) {
	response := serveAuthenticated(t, fakeNotificationService{err: errors.New("smtp mailpit:1025 failed")})
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", response.Code)
	}
	if response.Body.String() != "{\"code\":\"internal_error\",\"message\":\"internal server error\"}\n" {
		t.Fatalf("body leaks internal error: %q", response.Body.String())
	}
}

func serveAuthenticated(t *testing.T, service fakeNotificationService) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	New(service, fakeAuthService{userID: uuid.New()}).SetRoutes(mux)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/mock-notifications/energy", strings.NewReader(""))
	request.AddCookie(&http.Cookie{Name: "session_id", Value: uuid.NewString()})
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	return response
}
