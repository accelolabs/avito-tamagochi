package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	petmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/model"
	petservice "github.com/accelolabs/avito-tamagochi/backend/internal/game/pet/service"
	rewardmodel "github.com/accelolabs/avito-tamagochi/backend/internal/game/rewards/model"
	"github.com/google/uuid"
)

type fakeGameService struct{}

func (fakeGameService) GetPet(_ context.Context, _ uuid.UUID) (*petmodel.Stats, error) {
	return &petmodel.Stats{XP: 100, Level: 2, Stage: petmodel.Egg, Energy: 50, LastChargedAt: time.Unix(10, 0).UTC()}, nil
}

func TestAvailableRewardSerializesUsedAtAsNull(t *testing.T) {
	response := httptest.NewRecorder()
	writeJSON(response, http.StatusOK, toRewardResponse(rewardmodel.UserReward{
		ID: uuid.New(), Level: 2, RewardType: rewardmodel.Promotion, Status: "available", UnlockedAt: time.Unix(10, 0).UTC(),
	}))

	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	usedAt, exists := payload["usedAt"]
	if !exists || usedAt != nil {
		t.Fatalf("usedAt = %#v, exists = %v; want null field", usedAt, exists)
	}
}

func (fakeGameService) ChargePet(_ context.Context, _ uuid.UUID) (*petmodel.ChargeResult, error) {
	return &petmodel.ChargeResult{
		Pet:          &petmodel.Stats{XP: 102, Level: 2, Stage: petmodel.Egg, Energy: 70, LastChargedAt: time.Unix(20, 0).UTC()},
		BaseChargeXP: 2, DailyRewardXP: 10, TotalXPAwarded: 12,
	}, nil
}
func (fakeGameService) PetPet(_ context.Context, _ uuid.UUID) (*petmodel.PetActionResult, error) {
	return &petmodel.PetActionResult{Pet: &petmodel.Stats{Energy: 50, Level: 1, Stage: petmodel.Egg}, XPAwarded: 5}, nil
}

func (fakeGameService) GetStreak(_ context.Context, _ uuid.UUID) (*petmodel.StreakStats, error) {
	return &petmodel.StreakStats{NextDailyRewardXP: 10}, nil
}

var _ petservice.Service = fakeGameService{}

func TestPetResponseMapping(t *testing.T) {
	response := toPetResponse(&petmodel.Stats{XP: 100, Level: 2, Stage: petmodel.Egg, Energy: 50, LastChargedAt: time.Unix(10, 0).UTC()})
	if response.XP != 100 || response.Level != 2 || response.Stage != petmodel.Egg || response.Energy != 50 {
		t.Fatalf("unexpected pet response: %+v", response)
	}
}

func TestChargeResultResponseSeparatesAwards(t *testing.T) {
	response := httptest.NewRecorder()
	writeJSON(response, http.StatusOK, chargeResultResponse{
		Pet:          petResponse{XP: 12, Level: 1, Stage: petmodel.Egg, Energy: 70},
		BaseChargeXP: 2, DailyRewardXP: 10, TotalXPAwarded: 12,
	})

	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["baseChargeXp"] != float64(2) || payload["dailyRewardXp"] != float64(10) || payload["totalXpAwarded"] != float64(12) {
		t.Fatalf("unexpected charge award response: %#v", payload)
	}
	pet, ok := payload["pet"].(map[string]any)
	if !ok {
		t.Fatalf("pet = %#v, want object", payload["pet"])
	}
	if _, exists := pet["chargeStreak"]; exists {
		t.Fatalf("pet response still contains chargeStreak: %#v", pet)
	}
}

func TestPetActionResultSerializesAward(t *testing.T) {
	response := httptest.NewRecorder()
	writeJSON(response, http.StatusOK, petActionResultResponse{Pet: petResponse{Energy: 50}, XPAwarded: 5})
	if response.Body.String() != "{\"pet\":{\"xp\":0,\"level\":0,\"stage\":\"\",\"energy\":50,\"lastChargedAt\":\"0001-01-01T00:00:00Z\",\"isDead\":false},\"xpAwarded\":5}\n" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestChargeRejectsUnknownAction(t *testing.T) {
	handler := &Handler{}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/pet/actions", strings.NewReader(`"feed"`))
	response := httptest.NewRecorder()

	handler.applyPetAction(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestChargeRejectsTrailingJSONValue(t *testing.T) {
	handler := &Handler{}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/pet/actions", strings.NewReader(`"charge" {}`))
	response := httptest.NewRecorder()

	handler.applyPetAction(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestTaskEndpointRequiresAuthentication(t *testing.T) {
	handler := &Handler{}
	response := httptest.NewRecorder()

	handler.getTodayTasks(response, httptest.NewRequest(http.MethodGet, "/api/v1/tasks/today", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestLeaderboardSerializesCurrentUserAsNull(t *testing.T) {
	response := httptest.NewRecorder()
	writeJSON(response, http.StatusOK, leaderboardResponse[leaderboardEntryResponse]{
		Entries:     []leaderboardEntryResponse{},
		CurrentUser: nil,
	})

	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	currentUser, exists := payload["currentUser"]
	if !exists || currentUser != nil {
		t.Fatalf("currentUser = %#v, exists = %v; want null field present in JSON", currentUser, exists)
	}
}

func TestSummarySerializesEmptyUnlockedRewardsAsArray(t *testing.T) {
	response := httptest.NewRecorder()
	writeJSON(response, http.StatusOK, summaryResponse{
		UnlockedRewards: []string{},
	})

	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	unlocked, exists := payload["unlockedRewards"]
	if !exists {
		t.Fatalf("unlockedRewards key missing from JSON")
	}
	slice, ok := unlocked.([]any)
	if !ok || slice == nil {
		t.Fatalf("unlockedRewards = %#v; want empty JSON array", unlocked)
	}
}
