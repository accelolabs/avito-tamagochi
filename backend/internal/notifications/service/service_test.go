package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	notificationmodel "github.com/accelolabs/avito-tamagochi/backend/internal/notifications/model"
	"github.com/google/uuid"
)

type fakeRepository struct {
	acquired      bool
	participants  []notificationmodel.Participant
	delivered     map[uuid.UUID]map[int]bool
	processErrors map[uuid.UUID]error
}

func (r *fakeRepository) TryRunLock(context.Context) (func(), bool, error) {
	return func() {}, r.acquired, nil
}

func (r *fakeRepository) ListParticipants(context.Context) ([]notificationmodel.Participant, error) {
	return r.participants, nil
}

func (r *fakeRepository) WithParticipantLock(ctx context.Context, participant notificationmodel.Participant, handle notificationmodel.ParticipantHandler) (bool, error) {
	if err := r.processErrors[participant.UserID]; err != nil {
		return false, err
	}
	if r.delivered == nil {
		r.delivered = make(map[uuid.UUID]map[int]bool)
	}
	if r.delivered[participant.UserID] == nil {
		r.delivered[participant.UserID] = make(map[int]bool)
	}
	threshold, err := handle(ctx, participant, notificationmodel.DeliveredThresholds(r.delivered[participant.UserID]))
	if err != nil {
		return false, err
	}
	if threshold == nil {
		return false, nil
	}
	r.delivered[participant.UserID][*threshold] = true
	return true, nil
}

type recordingMailer struct {
	messages []notificationmodel.Message
	failures int
}

func (m *recordingMailer) Send(_ context.Context, message notificationmodel.Message) error {
	if m.failures > 0 {
		m.failures--
		return errors.New("smtp unavailable")
	}
	m.messages = append(m.messages, message)
	return nil
}

func TestDispatchAllSelectsCurrentThresholdAndDeduplicates(t *testing.T) {
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	energyValues := []int{100, 51, 50, 26, 25, 6, 5, 1, 0}
	repo := &fakeRepository{acquired: true}
	for index, energy := range energyValues {
		repo.participants = append(repo.participants, notificationmodel.Participant{
			UserID: uuid.New(), Email: fmt.Sprintf("user-%d@example.com", index),
			EnergyPercent: energy, EnergyUpdatedAt: now,
		})
	}
	mailer := &recordingMailer{}
	dispatcher := New(repo, mailer).(*service)
	dispatcher.now = func() time.Time { return now }

	first, err := dispatcher.DispatchAll(context.Background())
	if err != nil {
		t.Fatalf("first dispatch: %v", err)
	}
	if first.Participants != 9 || first.Sent != 7 || first.Skipped != 2 || first.Failed != 0 {
		t.Fatalf("first result = %+v", first)
	}
	second, err := dispatcher.DispatchAll(context.Background())
	if err != nil {
		t.Fatalf("second dispatch: %v", err)
	}
	if second.Sent != 0 || second.Skipped != 9 || len(mailer.messages) != 7 {
		t.Fatalf("deduplicated result = %+v messages=%d", second, len(mailer.messages))
	}
}

func TestDispatchAllRetriesSMTPFailureAndContinuesBatch(t *testing.T) {
	now := time.Date(2026, 8, 14, 12, 0, 0, 0, time.UTC)
	repo := &fakeRepository{acquired: true, participants: []notificationmodel.Participant{
		{UserID: uuid.New(), Email: "first@example.com", EnergyPercent: 25, EnergyUpdatedAt: now},
		{UserID: uuid.New(), Email: "second@example.com", EnergyPercent: 25, EnergyUpdatedAt: now},
	}}
	mailer := &recordingMailer{failures: 1}
	dispatcher := New(repo, mailer).(*service)
	dispatcher.now = func() time.Time { return now }

	first, err := dispatcher.DispatchAll(context.Background())
	if err == nil || first.Sent != 1 || first.Failed != 1 {
		t.Fatalf("first result = %+v error=%v", first, err)
	}
	second, err := dispatcher.DispatchAll(context.Background())
	if err != nil || second.Sent != 1 || second.Skipped != 1 || second.Failed != 0 {
		t.Fatalf("retry result = %+v error=%v", second, err)
	}
}

func TestDispatchAllSkipsWhenAnotherRunHoldsLock(t *testing.T) {
	repo := &fakeRepository{acquired: false, participants: []notificationmodel.Participant{{UserID: uuid.New()}}}
	result, err := New(repo, &recordingMailer{}).DispatchAll(context.Background())
	if err != nil || result != (notificationmodel.BatchResult{}) {
		t.Fatalf("result = %+v error=%v", result, err)
	}
}

func TestTemplatesKeepExactRussianCopy(t *testing.T) {
	want := map[int]string{
		50: "Я по тебе соскучился... 🥺\nНавестишь меня?",
		25: "Я что-то совсем без сил... 😥\nЭнергии почти не осталось. Заглянешь зарядить меня?",
		5:  "Кажется, я сейчас совсем сяду… 🪫\nКогда энергия закончится, я потеряю накопленный прогресс. Поможешь мне сохранить его?",
		0:  "Я полностью разрядился... 😭\nВесь мой накопленный опыт стирается. Надеюсь, мы сможем еще увидеться.",
	}
	for threshold, body := range want {
		selected := templates[threshold]
		if got := selected.firstLine + "\n" + selected.secondLine; got != body {
			t.Errorf("threshold %d body = %q", threshold, got)
		}
	}
}
