package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/auth/middleware"
	authservice "github.com/accelolabs/avito-tamagochi/backend/internal/auth/service"
	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
	notificationservice "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/service"
)

type Handler struct {
	service     notificationservice.Service
	authService authservice.Service
}

type dispatchResponse struct {
	Status    notificationmodel.Status `json:"status"`
	Energy    int                      `json:"energy"`
	Threshold *int                     `json:"threshold"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func New(service notificationservice.Service, authService authservice.Service) *Handler {
	return &Handler{service: service, authService: authService}
}

func (h *Handler) SetRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/mock-notifications/energy", middleware.RequireSession(h.authService, http.HandlerFunc(h.dispatchEnergy)))
}

func (h *Handler) dispatchEnergy(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication is required")
		return
	}
	result, err := h.service.DispatchEnergy(r.Context(), userID)
	if err != nil {
		log.Printf("dispatch energy notification: %v", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, dispatchResponse{Status: result.Status, Energy: result.Energy, Threshold: result.Threshold})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorResponse{Code: code, Message: message})
}
