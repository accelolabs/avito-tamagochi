package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth"
	"github.com/google/uuid"
)

type fakeService struct{ userID uuid.UUID }

func (f fakeService) Register(context.Context, auth.RegisterInput) (*auth.User, *auth.Session, error) {
	return nil, nil, nil
}
func (f fakeService) Login(context.Context, auth.LoginInput) (*auth.User, *auth.Session, error) {
	return nil, nil, nil
}
func (f fakeService) Logout(context.Context, uuid.UUID) error { return nil }
func (f fakeService) ValidateSession(context.Context, uuid.UUID) (uuid.UUID, error) {
	return f.userID, nil
}
func (f fakeService) FindUser(context.Context, uuid.UUID) (*auth.User, error) {
	return nil, sql.ErrNoRows
}

func TestRequireSessionAddsUserIDToContext(t *testing.T) {
	userID := uuid.New()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := UserID(r.Context())
		if !ok || got != userID {
			t.Fatalf("context user ID = %v, %v; want %v, true", got, ok, userID)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := RequireSession(fakeService{userID: userID}, next)
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.AddCookie(&http.Cookie{Name: "session_id", Value: uuid.NewString()})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestRequireSessionRejectsMissingCookie(t *testing.T) {
	handler := RequireSession(fakeService{}, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/protected", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
