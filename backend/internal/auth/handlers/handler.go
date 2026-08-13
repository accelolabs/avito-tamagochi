package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	autherrors "github.com/accelolabs/avito-tamagochi/backend/internal/auth/errors"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/model"
	authservice "github.com/accelolabs/avito-tamagochi/backend/internal/auth/service"
	"github.com/google/uuid"
)

const sessionCookieName = "session_id"

type Handler struct {
	service authservice.Service
	secure  bool
}

func New(service authservice.Service) *Handler {
	return &Handler{service: service, secure: os.Getenv("APP_SECURE_COOKIES") == "true"}
}

func (h *Handler) SetRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.register)
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.logout)
	mux.Handle("GET /api/v1/auth/me", middleware.RequireSession(h.service, http.HandlerFunc(h.Me)))
}

type userResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	CreatedAt   time.Time `json:"createdAt"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}

func toUserResponse(user *model.User) userResponse {
	return userResponse{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, CreatedAt: user.CreatedAt}
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, session *model.Session) {
	// #nosec G124 -- Secure is environment-controlled so local HTTP development remains usable; HttpOnly and SameSite are always enforced.
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID.String(),
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	// #nosec G124 -- Secure must match the environment used when the session cookie was created.
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(1, 0).UTC(), HttpOnly: true, Secure: h.secure, SameSite: http.SameSiteLaxMode})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", "request body is invalid")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "validation_error", "request body is invalid")
		return false
	}
	return true
}

func decodeStringObject(data []byte, expectedKeys ...string) (map[string]string, error) {
	var values map[string]string
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, err
	}
	if len(values) != len(expectedKeys) {
		return nil, errors.New("unexpected JSON fields")
	}
	for _, key := range expectedKeys {
		if _, ok := values[key]; !ok {
			return nil, errors.New("required JSON field is missing")
		}
	}
	return values, nil
}

func mapServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, autherrors.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "validation_error", "request is invalid")
	case errors.Is(err, autherrors.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, "email_already_exists", "email is already registered")
	case errors.Is(err, autherrors.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
