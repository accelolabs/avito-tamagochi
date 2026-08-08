package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	petservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/service"
	"github.com/google/uuid"
)

type fakeGameService struct{}

func (fakeGameService) GetPet(_ context.Context, _ uuid.UUID) (*petmodel.Stats, error) {
	return &petmodel.Stats{XP: 100, Level: 2, Energy: 50, LastChargedAt: time.Unix(10, 0).UTC()}, nil
}

func (fakeGameService) ChargePet(_ context.Context, _ uuid.UUID) (*petmodel.Stats, error) {
	return &petmodel.Stats{XP: 110, Level: 2, Energy: 100, LastChargedAt: time.Unix(20, 0).UTC()}, nil
}

var _ petservice.Service = fakeGameService{}

func TestPetResponseMapping(t *testing.T) {
	response := toPetResponse(&petmodel.Stats{XP: 100, Level: 2, Energy: 50, LastChargedAt: time.Unix(10, 0).UTC()})
	if response.XP != 100 || response.Level != 2 || response.Energy != 50 {
		t.Fatalf("unexpected pet response: %+v", response)
	}
}

func TestChargeRejectsUnknownAction(t *testing.T) {
	handler := &Handler{}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/pet/actions", strings.NewReader(`"feed"`))
	response := httptest.NewRecorder()

	handler.chargePet(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestUnimplementedTaskEndpointReturnsNotImplemented(t *testing.T) {
	handler := &Handler{}
	response := httptest.NewRecorder()

	handler.getTodayTasks(response, httptest.NewRequest(http.MethodGet, "/api/v1/tasks/today", nil))

	if response.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotImplemented)
	}
}
